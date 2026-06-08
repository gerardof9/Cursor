#!/usr/bin/ksh
#
# Formatea el archivo consolidado de particiones Oracle para envio por correo.
# Entrada: archivo consolidado (argumento 1). Salida: stdout unicamente.
#
#

INPUT_FILE=$1

if [ -z "$INPUT_FILE" ]; then
	echo "Error: falta argumento con archivo consolidado" >&2
	exit 1
fi

if [ ! -f "$INPUT_FILE" ]; then
	echo "Error: archivo no encontrado: $INPUT_FILE" >&2
	exit 1
fi

if [ ! -s "$INPUT_FILE" ]; then
	echo "Advertencia: archivo consolidado vacio: $INPUT_FILE" >&2
fi

export LC_TIME=C
FECHA=$(date +"%d-%b-%Y")

awk -v fecha="$FECHA" '
BEGIN {
	crit_n = 0
	alert_n = 0
	FS = "\034"
}
function sort_array(arr, n,    i, j, tmp) {
	for (i = 1; i < n; i++) {
		for (j = i + 1; j <= n; j++) {
			if (arr[i] > arr[j]) {
				tmp = arr[i]
				arr[i] = arr[j]
				arr[j] = tmp
			}
		}
	}
}

function trim(s) {
	sub(/^[ \t\r\n]+/, "", s)
	sub(/[ \t\r\n]+$/, "", s)
	return s
}

function print_section(title, subtitle) {
	print "===================================================================="
	print title
	print "===================================================================="
	print ""
	if (subtitle != "") {
		print subtitle
		print ""
	}
}

/^=== / {
	host = trim($0)
	sub(/^=== /, "", host)
	sub(/ ===$/, "", host)
	next
}

/^-- / {
	sid = trim($0)
	sub(/^-- /, "", sid)
	sub(/ --$/, "", sid)
	next
}

$NF ~ /^[0-9]+\.[0-9]+$/ && $1 != "OWNER" && $1 !~ /^-+/ {
	meses = $NF + 0
	owner = $1
	table = $2
	key = sprintf("%8.1f", meses) FS host FS sid FS owner FS table

	if (meses <= 1) {
		crit[++crit_n] = key
	} else if (meses <= 3) {
		alert[++alert_n] = key
	} else {
		ok_key = host FS sid
		ok_count[ok_key]++
		if (!(ok_key in ok_min) || meses < ok_min[ok_key]) {
			ok_min[ok_key] = meses
		}
		ok_host[ok_key] = host
		ok_sid[ok_key] = sid
	}
	next
}

END {
	print_section("REPORTE DE PARTICIONES ORACLE", "")
	print "Fecha ejecucion : " fecha
	print ""
	print "Criterios de clasificacion:"
	print "Ventana critica : <= 1 mes restante"
	print "Ventana alerta  : <= 3 meses restantes"
	print ""

	print_section("PARTICIONES CRITICAS (<= 1 MES)", "Total tablas criticas: " crit_n)
	if (crit_n > 0) {
		printf "%-14s %-12s %-18s %-28s %s\n", "SERVIDOR", "INSTANCIA", "OWNER", "TABLE_NAME", "MESES"
		sort_array(crit, crit_n)
		for (i = 1; i <= crit_n; i++) {
			n = split(crit[i], f, FS)
			if (n >= 5) {
				printf "%-14s %-12s %-18s %-28s %.1f\n", f[2], f[3], f[4], f[5], f[1] + 0
			}
		}
	}
	print ""

	print_section("PARTICIONES EN ALERTA (> 1 Y <= 3 MESES)", "Total tablas en alerta: " alert_n)
	if (alert_n > 0) {
		printf "%-14s %-12s %-18s %-28s %s\n", "SERVIDOR", "INSTANCIA", "OWNER", "TABLE_NAME", "MESES"
		sort_array(alert, alert_n)
		for (i = 1; i <= alert_n; i++) {
			n = split(alert[i], f, FS)
			if (n >= 5) {
				printf "%-14s %-12s %-18s %-28s %.1f\n", f[2], f[3], f[4], f[5], f[1] + 0
			}
		}
	}
	print ""

	print_section("TABLAS SIN RIESGO INMEDIATO (> 3 MESES)", "")
	printf "%-14s %-12s %-11s %s\n", "SERVIDOR", "INSTANCIA", "TABLAS_OK", "MENOR_HORIZONTE"
	ok_n = 0
	for (k in ok_count) {
		ok[++ok_n] = ok_host[k] FS ok_sid[k] FS ok_count[k] FS ok_min[k]
	}
	sort_array(ok, ok_n)
	for (i = 1; i <= ok_n; i++) {
		n = split(ok[i], f, FS)
		if (n >= 4) {
			printf "%-14s %-12s %-11s minimo %.1f meses\n", f[1], f[2], f[3], f[4]
		}
	}
}
' "$INPUT_FILE"

AWK_RC=$?
if [ $AWK_RC -ne 0 ]; then
	echo ""
	echo "Error: fallo el formateo awk (codigo $AWK_RC)"
fi

echo ""
echo "===================================================================="
echo "DETALLE COMPLETO"
echo "===================================================================="
echo ""
cat "$INPUT_FILE"
