#!/bin/sh
#
# Reporte de database links en instancias de desarrollo que apuntan a producción.
# Recorrido de instancias/PDB: mismo criterio que dbsat.sh
#

export JAVA_HOME=/usr/java8_64

ORATAB=/etc/oratab
export ORATAB
HOME=/home/oracle
PROD_FILE=${HOME}/ora_inst_prod.txt

REMOTE_USER=ususopbd
REMOTE_HOST=dwhsrvp01
REMOTE_DIR=/ora_scripts/logs/dblinks

HOST_SHORT=`hostname | sed 's/\..*//'`
LOGFILE=${HOME}/dblinks_${HOST_SHORT}.txt
HTMLFILE=${HOME}/dblinks_${HOST_SHORT}.html
SUMMARY_TMP=${HOME}/.dblinks_${HOST_SHORT}.$$
TXT_TMP=${HOME}/.dblinks_txt_${HOST_SHORT}.$$


get_oracle_home()
{
    awk -F: -v sid="$1" '
    tolower($1) == tolower(sid) {
        print $2
        exit
    }' "${ORATAB}"
}

get_pdbs()
{
sqlplus -s "/ as sysdba" <<EOF
set pages 0
set heading off
set feedback off
set verify off
set trimspool on
set lines 200
select name
from v\$containers
where name not in ('CDB\$ROOT','PDB\$SEED')
and open_mode='READ WRITE';
exit
EOF
}

# DB_LINK y destino (BD) desde DBA_DB_LINKS.HOST
get_db_links()
{
    _container="$1"

    if [ -n "${_container}" ]; then
        unset ORACLE_PDB_SID
    fi

    if [ -n "${_container}" ]; then
        _alter="alter session set container = ${_container};"
    else
        _alter=""
    fi

sqlplus -s "/ as sysdba" <<EOF
set pages 0
set heading off
set feedback off
set verify off
set trimspool on
set lines 200
${_alter}
select upper(db_link) || '|' || upper(bd) from (
  select db_link,
    regexp_replace(
      regexp_substr(upper(host) ,'(SERVICE_NAME|SID)\\s*=\\s*(\\S+\\.(\\w+)|\\w+)|^\\w+'),
      '(SERVICE_NAME|SID)\\s*=\\s*','') bd
  from dba_db_links
  where host is not null
)
where bd is not null;
exit
EOF
}

# Coincidencia exacta en ora_inst_prod.txt; si termina en SVC y no matchea, probar sin SVC
match_prod_entry()
{
    _t="$1"

    if grep -Fx "${_t}" "${PROD_FILE}" >/dev/null 2>&1; then
        echo "${_t}"
        return 0
    fi

    case "${_t}" in
        *SVC)
            _base=`echo "${_t}" | sed 's/SVC$//'`
            if grep -Fx "${_base}" "${PROD_FILE}" >/dev/null 2>&1; then
                echo "${_base}"
                return 0
            fi
            ;;
    esac

    return 1
}

write_host_report()
{
    sort -t'|' -k1,1 "${TXT_TMP}" | awk -F'|' '
    {
        printf "%-40s %-16s %-12s %s\n", $1, $2, $3, $4
    }' >> "${LOGFILE}"
}

html_spans()
{
    _html=""
    for _n in $1; do
        _html="${_html}<span>${_n}</span>"
    done
    echo "${_html}"
}

html_write_head()
{
cat >> "${HTMLFILE}" <<EOF
<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>DB Links - ${HOST_SHORT}</title>
<style>
  body {
    font-family: system-ui, -apple-system, "Segoe UI", sans-serif;
    font-size: 14px;
    color: #333;
    background: #fafbfc;
    margin: 2rem;
    line-height: 1.35;
  }
  .report-wrap {
    max-width: 960px;
  }
  .page {
    box-sizing: border-box;
    width: 100%;
    background: #fff;
    padding: 1.5rem 1.75rem;
    border: 1px solid #e8eef3;
    margin-bottom: 1.5rem;
  }
  .report-title {
    box-sizing: border-box;
    width: 100%;
    background: #e8f4fc;
    padding: 0.75rem 1.75rem;
    margin: 0 0 1rem;
    border: 1px solid #e8eef3;
    border-bottom: 1px solid #d6e9f5;
    text-align: center;
  }
  .report-title h1 {
    font-size: 1.25rem;
    font-weight: 600;
    margin: 0;
    color: #1a4a6e;
  }
  .title-block {
    background: #dceef8;
    padding: 0.75rem 1rem;
    margin: -1.5rem -1.75rem 1rem;
    border-bottom: 1px solid #c5dff0;
  }
  .host {
    color: #4a6a82;
    margin: 0;
    font-size: 0.9rem;
  }
  .host strong {
    font-weight: 700;
    color: #1a4a6e;
  }
  h2 {
    font-size: 0.85rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #3d6a8a;
    margin: 1.25rem 0 0.5rem;
    padding: 0.35rem 0.65rem;
    background: #eef7fc;
    border-left: 3px solid #b8d9ed;
  }
  h2:first-of-type { margin-top: 0; }
  table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: 0.6rem;
  }
  th {
    text-align: left;
    font-size: 0.8rem;
    font-weight: 600;
    color: #4a6578;
    padding: 0.3rem 0.5rem;
    background: #f0f8fc;
    border-bottom: 1px solid #d6e9f5;
  }
  td {
    padding: 0.3rem 0.5rem;
    vertical-align: top;
    border-bottom: 1px solid #eef2f5;
    line-height: 1.3;
  }
  td:first-child {
    width: 140px;
    font-family: ui-monospace, Consolas, monospace;
    font-size: 0.9rem;
    color: #111;
    background: #fafcfd;
  }
  .list {
    font-family: ui-monospace, Consolas, monospace;
    font-size: 0.9rem;
    color: #444;
    line-height: 1.3;
  }
  .list span:not(:last-child)::after {
    content: ", ";
    color: #999;
  }
  .empty-msg {
    color: #666;
    font-size: 0.9rem;
    margin: 0.5rem 0;
  }
</style>
</head>
<body>
EOF
}

html_write_foot()
{
    echo '</body>' >> "${HTMLFILE}"
    echo '</html>' >> "${HTMLFILE}"
}

write_host_report_html()
{
    : > "${HTMLFILE}"
    html_write_head

    echo '<div class="report-wrap">' >> "${HTMLFILE}"
    echo '<div class="report-title"><h1>DB Links Desarrollo a Produccion</h1></div>' >> "${HTMLFILE}"
    echo '<div class="page">' >> "${HTMLFILE}"
    echo '  <div class="title-block">' >> "${HTMLFILE}"
    echo "    <p class=\"host\">HOST: <strong>${HOST_SHORT}</strong></p>" >> "${HTMLFILE}"
    echo '  </div>' >> "${HTMLFILE}"

    echo '  <h2>Por base no productiva</h2>' >> "${HTMLFILE}"
    echo '  <table>' >> "${HTMLFILE}"
    echo '    <thead><tr><th>No productiva</th><th>Apunta a</th></tr></thead>' >> "${HTMLFILE}"
    echo '    <tbody>' >> "${HTMLFILE}"

    sort "${SUMMARY_TMP}" |
    while IFS='|' read -r _inst _prods
    do
        _list=`html_spans "${_prods}"`
        echo "      <tr><td>${_inst}</td><td class=\"list\">${_list}</td></tr>" >> "${HTMLFILE}"
    done

    echo '    </tbody>' >> "${HTMLFILE}"
    echo '  </table>' >> "${HTMLFILE}"

    echo '  <h2>Por base de produccion</h2>' >> "${HTMLFILE}"
    echo '  <table>' >> "${HTMLFILE}"
    echo '    <thead><tr><th>BD Producción</th><th>Apuntada por</th></tr></thead>' >> "${HTMLFILE}"
    echo '    <tbody>' >> "${HTMLFILE}"

    awk -F'|' '
    {
        inst = $1
        n = split($2, p, " ")
        for (i = 1; i <= n; i++) {
            if (p[i] == "") continue
            c = ++ic[p[i]]
            pi[p[i], c] = inst
        }
    }
    END {
        for (prod in ic) print prod
    }' "${SUMMARY_TMP}" | sort |
    while read _prod
    do
        _insts=`awk -F'|' -v prod="${_prod}" '
        {
            n = split($2, p, " ")
            for (i = 1; i <= n; i++)
                if (p[i] == prod) print $1
        }' "${SUMMARY_TMP}" | sort -u | tr '\n' ' ' | sed 's/ $//'`
        _list=`html_spans "${_insts}"`
        echo "      <tr><td>${_prod}</td><td class=\"list\">${_list}</td></tr>" >> "${HTMLFILE}"
    done

    echo '    </tbody>' >> "${HTMLFILE}"
    echo '  </table>' >> "${HTMLFILE}"
    echo '</div>' >> "${HTMLFILE}"
    echo '</div>' >> "${HTMLFILE}"
    html_write_foot
}

process_instance()
{
    _name=`echo "$1" | tr 'a-z' 'A-Z'`
    _container="$2"
    _prod_line=""

    while IFS='|' read -r _link _bd
    do
        _prod=`match_prod_entry "${_bd}"`
        if [ -z "${_prod}" ]; then
            continue
        fi

        echo "${_link}|${HOST_SHORT}|${_name}|${_prod}" >> "${TXT_TMP}"

        case " ${_prod_line} " in
            *" ${_prod} "*) ;;
            *) _prod_line="${_prod_line} ${_prod}" ;;
        esac
    done <<EOF
`get_db_links "${_container}"`
EOF

    if [ -n "${_prod_line}" ]; then
        echo "${_name}|${_prod_line# }" >> "${SUMMARY_TMP}"
    fi
}


cd "${HOME}" || exit 1

: > "${SUMMARY_TMP}"
: > "${TXT_TMP}"

for PMON in `ps -ef | awk '$NF ~ /^ora_pmon_/ {print $NF}'`
do
    INSTANCE_SID=`echo "${PMON}" | awk -F_ '{print $3}'`

    export ORACLE_SID="${INSTANCE_SID}"
    unset ORACLE_PDB_SID

    ORACLE_HOME=`get_oracle_home "${INSTANCE_SID}"`

    if [ -z "${ORACLE_HOME}" ]; then
        continue
    fi

    export ORACLE_HOME
    export PATH=${ORACLE_HOME}/bin:${PATH}

    PDB_LIST=`get_pdbs | grep -E '^[A-Za-z][A-Za-z0-9_$#]*$'`

    if [ -n "${PDB_LIST}" ]; then
        for PDB in ${PDB_LIST}
        do
            export ORACLE_PDB_SID="${PDB}"
            process_instance "${PDB}" "${PDB}"
            unset ORACLE_PDB_SID
        done
    else
        unset ORACLE_PDB_SID
        process_instance "${INSTANCE_SID}" ""
    fi
done

if [ -s "${TXT_TMP}" ]; then
    : > "${LOGFILE}"
    write_host_report
    write_host_report_html
    scp "${LOGFILE}" ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/
    scp "${HTMLFILE}" ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/
fi

rm -f "${SUMMARY_TMP}" "${TXT_TMP}"
