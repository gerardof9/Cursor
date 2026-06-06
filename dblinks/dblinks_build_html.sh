#!/bin/sh
#
# Genera dblinks.html consolidado a partir de dblinks_<host>.txt
# Uso: dblinks_build_html.sh [directorio]
#

DIR="${1:-.}"
HEAD="${DIR}/dblinks_html_head.inc"
FOOT="${DIR}/dblinks_html_foot.inc"
OUT="${DIR}/dblinks.html"

cd "${DIR}" || exit 1

ROWS=${TMPDIR:-/tmp}/dblinks_html_rows.$$
trap 'rm -f "${ROWS}"' 0 1 2 3 15

: > "${ROWS}"

for _f in dblinks_*.txt; do
    [ -f "${_f}" ] || continue
    case "${_f}" in
        dblinks.txt) continue ;;
    esac

    awk '
    length($0) < 1 { next }
    {
        dblink = substr($0, 1, 40)
        host   = substr($0, 41, 16)
        desa   = substr($0, 57, 12)
        prod   = substr($0, 69)
        gsub(/[ \t\r]+$/, "", dblink)
        gsub(/^[ \t\r]+|[ \t\r]+$/, "", host)
        gsub(/^[ \t\r]+|[ \t\r]+$/, "", desa)
        gsub(/^[ \t\r]+/, "", prod)
        if (host != "" && desa != "" && prod != "")
            print host "|" desa "|" prod
    }' "${_f}" >> "${ROWS}"
done

if [ ! -s "${ROWS}" ]; then
    exit 0
fi

sort -u "${ROWS}" > "${ROWS}.sorted"
mv "${ROWS}.sorted" "${ROWS}"

html_table_header()
{
    _col1="$1"
    _col2="$2"
    echo '<table border="1" cellspacing="0" cellpadding="4" width="100%" style="border-collapse:collapse;font-family:Consolas,monospace;font-size:13px;">'
    echo '<colgroup><col width="140" style="width:140px"><col style="width:480px"></colgroup>'
    echo '<tr>'
    echo "<th align=\"left\" style=\"padding:4px 8px;background:#f0f8fc;font-size:12px;font-weight:bold;\">${_col1}</th>"
    echo "<th align=\"left\" style=\"padding:4px 8px;background:#f0f8fc;font-size:12px;font-weight:bold;\">${_col2}</th>"
    echo '</tr>'
}

html_table_row()
{
    _col1="$1"
    _col2="$2"
    echo "<tr><td valign=\"top\" style=\"padding:4px 8px;font-size:13px;background:#fafcfd;\">${_col1}</td><td valign=\"top\" style=\"padding:4px 8px;font-size:13px;\">${_col2}</td></tr>"
}

html_summary()
{
    echo '<div class="page" style="margin-bottom:10px;">'
    echo '<table border="0" cellspacing="0" cellpadding="0" width="100%" style="width:100%;border-collapse:collapse;border:1px solid #e8eef3;">'
    echo '<tr><td style="background:#eef7fc;padding:6px 8px;font-weight:bold;font-size:12px;color:#3d6a8a;border-bottom:1px solid #d6e9f5;">Resumen de accesos</td></tr>'
    echo '<tr><td style="padding:8px;">'
    html_table_header "BD Producción" "Apuntada por"

    awk -F'|' '
    {
        key = $3 SUBSEP $2
        if (!(key in seen)) {
            seen[key] = 1
            c = ++cnt[$3]
            di[$3, c] = $2
        }
    }
    END {
        for (prod in cnt) print prod
    }' "${ROWS}" | sort |
    while read _prod
    do
        _insts=`awk -F'|' -v prod="${_prod}" '
        {
            if ($3 == prod) print $2
        }' "${ROWS}" | sort -u | tr '\n' ',' | sed 's/,$//' | sed 's/,/, /g'`
        html_table_row "${_prod}" "${_insts}"
    done

    echo '</table>'
    echo '</td></tr>'
    echo '</table>'
    echo '</div>'
}

html_host_block()
{
    _host="$1"

    echo '<div class="page" style="margin-bottom:10px;">'
    echo '<table border="0" cellspacing="0" cellpadding="0" width="100%" style="width:100%;border-collapse:collapse;border:1px solid #e8eef3;">'
    echo "  <tr><td style=\"background:#dceef8;padding:6px 8px;font-size:12px;color:#4a6a82;border-bottom:1px solid #c5dff0;border-top:2px solid #c5dff0;\">HOST: <strong style=\"color:#1a4a6e;\">${_host}</strong></td></tr>"
    echo '  <tr><td style="padding:8px;">'

    echo '  <p style="font-size:12px;font-weight:bold;color:#3d6a8a;background:#eef7fc;padding:4px 8px;margin:0 0 4px;border-left:3px solid #b8d9ed;">Por base no productiva</p>'
    html_table_header "No productiva" "Apunta a"

    awk -F'|' -v host="${_host}" '
    $1 == host {
        key = $2 SUBSEP $3
        if (!(key in seen)) {
            seen[key] = 1
            print $2
        }
    }' "${ROWS}" | sort -u |
    while read _desa
    do
        _prods=`awk -F'|' -v host="${_host}" -v desa="${_desa}" '
        $1 == host && $2 == desa { print $3 }' "${ROWS}" | sort -u | tr '\n' ',' | sed 's/,$//' | sed 's/,/, /g'`
        html_table_row "${_desa}" "${_prods}"
    done

    echo '</table>'

    echo '  <p style="font-size:12px;font-weight:bold;color:#3d6a8a;background:#eef7fc;padding:4px 8px;margin:10px 0 4px;border-left:3px solid #b8d9ed;">Por base de produccion</p>'
    html_table_header "BD Producción" "Apuntada por"

    awk -F'|' -v host="${_host}" '
    $1 == host {
        key = $3 SUBSEP $2
        if (!(key in seen)) {
            seen[key] = 1
            print $3
        }
    }' "${ROWS}" | sort -u |
    while read _prod
    do
        _desas=`awk -F'|' -v host="${_host}" -v prod="${_prod}" '
        $1 == host && $3 == prod { print $2 }' "${ROWS}" | sort -u | tr '\n' ',' | sed 's/,$//' | sed 's/,/, /g'`
        html_table_row "${_prod}" "${_desas}"
    done

    echo '</table>'
    echo '  </td></tr>'
    echo '  </table>'
    echo '</div>'
}

cat "${HEAD}" > "${OUT}"
html_summary >> "${OUT}"

awk -F'|' '{ print $1 }' "${ROWS}" | sort -u |
while read _host
do
    html_host_block "${_host}"
done >> "${OUT}"

cat "${FOOT}" >> "${OUT}"
