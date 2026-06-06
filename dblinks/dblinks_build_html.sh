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

html_spans()
{
    _html=""
    for _n in $1; do
        _html="${_html}<span>${_n}</span>"
    done
    echo "${_html}"
}

html_summary()
{
    echo '<div class="page">'
    echo '  <h2>Resumen de accesos</h2>'
    echo '  <table>'
    echo '    <thead><tr><th>BD Producción</th><th>Apuntada por</th></tr></thead>'
    echo '  <tbody>'

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
        }' "${ROWS}" | sort -u | tr '\n' ' ' | sed 's/ $//'`
        _list=`html_spans "${_insts}"`
        echo "      <tr><td>${_prod}</td><td class=\"list\">${_list}</td></tr>"
    done

    echo '    </tbody>'
    echo '  </table>'
    echo '</div>'
}

html_host_block()
{
    _host="$1"

    echo '<div class="page">'
    echo '  <div class="title-block">'
    echo "    <p class=\"host\">HOST: <strong>${_host}</strong></p>"
    echo '  </div>'

    echo '  <h2>Por base no productiva</h2>'
    echo '  <table>'
    echo '    <thead><tr><th>No productiva</th><th>Apunta a</th></tr></thead>'
    echo '    <tbody>'

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
        $1 == host && $2 == desa { print $3 }' "${ROWS}" | sort -u | tr '\n' ' ' | sed 's/ $//'`
        _list=`html_spans "${_prods}"`
        echo "      <tr><td>${_desa}</td><td class=\"list\">${_list}</td></tr>"
    done

    echo '    </tbody>'
    echo '  </table>'

    echo '  <h2>Por base de produccion</h2>'
    echo '  <table>'
    echo '    <thead><tr><th>BD Producción</th><th>Apuntada por</th></tr></thead>'
    echo '    <tbody>'

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
        $1 == host && $3 == prod { print $2 }' "${ROWS}" | sort -u | tr '\n' ' ' | sed 's/ $//'`
        _list=`html_spans "${_desas}"`
        echo "      <tr><td>${_prod}</td><td class=\"list\">${_list}</td></tr>"
    done

    echo '    </tbody>'
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
