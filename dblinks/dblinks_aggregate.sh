#!/bin/sh
#
# Resumen agregado HTML: todas las BD prod del informe y las BD dev que apuntan.
# Lee dblinks_<host>.html (seccion "Por base no productiva").
# Uso: dblinks_aggregate.sh [directorio]
#

DIR="${1:-.}"

cd "${DIR}" || exit 1

AGG_PAIRS=${TMPDIR:-/tmp}/dblinks_agg_pairs.$$
trap 'rm -f "${AGG_PAIRS}"' 0 1 2 3 15

: > "${AGG_PAIRS}"

for _f in dblinks_*.html; do
    [ -f "${_f}" ] || continue
    case "${_f}" in
        dblinks.html) continue ;;
    esac

    awk '
    /<h2>Por base no productiva<\/h2>/ { s = 1; next }
    /<h2>Por base de produccion<\/h2>/ { s = 0; next }
    s && /<tr><td>/ {
        line = $0
        sub(/^.*<tr><td>/, "", line)
        n = index(line, "</td>")
        inst = substr(line, 1, n - 1)
        rest = substr(line, n + 5)
        sub(/^<td class="list">/, "", rest)
        sub(/<\/td><\/tr>.*$/, "", rest)
        gsub(/<\/span><span>/, "\n", rest)
        gsub(/<span>/, "", rest)
        gsub(/<\/span>/, "", rest)
        np = split(rest, prods, "\n")
        for (i = 1; i <= np; i++) {
            if (prods[i] != "") print prods[i] "|" inst
        }
    }' "${_f}" >> "${AGG_PAIRS}"
done

if [ ! -s "${AGG_PAIRS}" ]; then
    exit 0
fi

sort -u "${AGG_PAIRS}" | sort -t'|' -k1,1 -k2,2 > "${AGG_PAIRS}.sorted"
mv "${AGG_PAIRS}.sorted" "${AGG_PAIRS}"

html_spans()
{
    _html=""
    for _n in $1; do
        _html="${_html}<span>${_n}</span>"
    done
    echo "${_html}"
}

echo '<div class="page">'
echo '  <h2>Resumen de accesos</h2>'
echo '  <table>'
echo '    <thead><tr><th>BD Producción</th><th>Apuntada por</th></tr></thead>'
echo '  <tbody>'

awk -F'|' '
{
    if ($1 != prod) {
        if (prod != "") print prod "\t" insts
        prod = $1
        insts = $2
    } else {
        insts = insts " " $2
    }
}
END {
    if (prod != "") print prod "\t" insts
}' "${AGG_PAIRS}" |
while IFS='	' read -r _prod _insts
do
    _list=`html_spans "${_insts}"`
    echo "      <tr><td>${_prod}</td><td class=\"list\">${_list}</td></tr>"
done

echo '    </tbody>'
echo '  </table>'
echo '</div>'
