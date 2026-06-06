#!/bin/sh
#
# Envia dblinks.html como adjunto via sendmail (multipart/mixed).
# Uso: dblinks_send_mail.sh [directorio]
#

DIR="${1:-.}"
MAIL_FROM="${MAIL_FROM:-BaseDeDatos@aplicaciones.ute}"
MAIL_TO="${MAIL_TO:-SopBD@ute.com.uy}"
MAIL_SUBJECT="${MAIL_SUBJECT:-DB Links Desarrollo a Produccion}"
ATTACH="${ATTACH:-dblinks.html}"

cd "${DIR}" || exit 1

if [ ! -s "${ATTACH}" ]; then
    echo "No existe o esta vacio: ${ATTACH}" >&2
    exit 1
fi

B64=${TMPDIR:-/tmp}/dblinks_b64.$$
BOUNDARY="dblinks_${RANDOM:-$$}_$$"
trap 'rm -f "${B64}"' 0 1 2 3 15

if openssl base64 -e -in "${ATTACH}" -out "${B64}" 2>/dev/null; then
    :
elif openssl base64 "${ATTACH}" > "${B64}" 2>/dev/null; then
    :
elif uuencode "${ATTACH}" "${ATTACH}" 2>/dev/null | sed '1d;$d' > "${B64}"; then
    :
else
    echo "No se pudo codificar ${ATTACH} (openssl/uuencode)" >&2
    exit 1
fi

{
    echo "From: ${MAIL_FROM}"
    echo "To: ${MAIL_TO}"
    echo "Subject: ${MAIL_SUBJECT}"
    echo "MIME-Version: 1.0"
    echo "Content-Type: multipart/mixed; boundary=\"${BOUNDARY}\""
    echo ""
    echo "--${BOUNDARY}"
    echo "Content-Type: text/plain; charset=UTF-8"
    echo ""
    echo "Reporte DB Links Desarrollo a Produccion (adjunto)."
    echo ""
    echo "--${BOUNDARY}"
    echo "Content-Type: text/html; charset=UTF-8"
    echo "Content-Disposition: attachment; filename=\"${ATTACH}\""
    echo "Content-Transfer-Encoding: base64"
    echo ""
    cat "${B64}"
    echo ""
    echo "--${BOUNDARY}--"
} | /usr/lib/sendmail -t
