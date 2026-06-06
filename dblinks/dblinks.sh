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

process_instance()
{
    _name=`echo "$1" | tr 'a-z' 'A-Z'`
    _container="$2"

    while IFS='|' read -r _link _bd
    do
        _prod=`match_prod_entry "${_bd}"`
        if [ -z "${_prod}" ]; then
            continue
        fi

        echo "${_link}|${HOST_SHORT}|${_name}|${_prod}" >> "${TXT_TMP}"
    done <<EOF
`get_db_links "${_container}"`
EOF
}


cd "${HOME}" || exit 1

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
    scp "${LOGFILE}" ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/
fi

rm -f "${TXT_TMP}"
