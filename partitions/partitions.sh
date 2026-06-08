#!/usr/bin/ksh
. /home/oracle/.profile

function runQuery {
	export ORACLE_HOME=${1}
	export ORACLE_SID=${2} 
	export SQLOUTPUT=`$ORACLE_HOME/bin/sqlplus -s / as sysdba <<EOF
                   
	SET HEADING ON
    SET PAGESIZE 1000
	SET FEEDBACK OFF
    SET VERIFY OFF
    SET ECHO OFF
    SET TRIMSPOOL ON
    SET TERMOUT OFF
    SET LINESIZE 200
	SET TAB OFF

	COLUMN owner FORMAT A18
	COLUMN table_name FORMAT A35
	COLUMN ultima_particion FORMAT A25
	COLUMN valor_limite FORMAT A15
	COLUMN meses_restantes FORMAT A15
	
	COLUMN owner            HEADING 'OWNER'
	COLUMN table_name       HEADING 'TABLE_NAME'
	COLUMN ultima_particion HEADING 'ULTIMA_PARTICION'
	COLUMN valor_limite     HEADING 'VALOR_LIMITE'
	COLUMN meses_restantes  HEADING 'MESES_RESTANTES'

	WITH raw_partitions AS (
    SELECT /*+ MATERIALIZE */
        tp.table_owner, tp.table_name, tp.partition_name, tp.partition_position, tc.data_type, tp.high_value,
        CASE WHEN pt.interval IS NOT NULL THEN 'YES' ELSE 'NO' END as auto_type
    FROM dba_tab_partitions tp
    JOIN dba_part_tables pt ON tp.table_owner = pt.owner AND tp.table_name = pt.table_name
    JOIN dba_part_key_columns pk ON tp.table_owner = pk.owner AND tp.table_name = pk.name
    JOIN dba_tab_columns tc ON pk.owner = tc.owner AND pk.name = tc.table_name AND pk.column_name = tc.column_name
    WHERE pt.partitioning_type = 'RANGE'
      AND tp.table_owner NOT IN ('SYS','SYSTEM','SYSMAN','AUDSYS','MDSYS','WMSYS','XDB','CTXSYS','DBSNMP','OUTLN','GSMADMIN_INTERNAL')
      AND (tc.data_type IN ('DATE', 'VARCHAR2', 'CHAR') OR tc.data_type LIKE 'TIMESTAMP%')
),
evaluated_partitions AS (
    SELECT table_owner, table_name, partition_name, partition_position, data_type, xml_hv, auto_type
    FROM (
        SELECT r.*,
               CASE WHEN ROW_NUMBER() OVER (PARTITION BY table_owner, table_name ORDER BY partition_position DESC) <= 5 
                    THEN dbms_xmlgen.getxml('SELECT high_value FROM dba_tab_partitions WHERE table_owner='''||table_owner||''' AND table_name='''||table_name||''' AND partition_name='''||partition_name||'''')
                    ELSE NULL END as xml_hv
        FROM raw_partitions r
    )
    WHERE xml_hv IS NOT NULL
),
final_dates AS (
    SELECT table_owner, table_name, partition_name, data_type, auto_type,
        COALESCE(
            regexp_substr(xml_hv, '[0-9]{4}-[0-9]{2}-[0-9]{2}'),
            regexp_substr(xml_hv, '[0-9]{2}-[A-Z,a-z,0-9]{3}-[0-9]{4}')
        ) as fecha_str,
        ROW_NUMBER() OVER (PARTITION BY table_owner, table_name ORDER BY partition_position DESC) as rn
    FROM evaluated_partitions
    WHERE xml_hv NOT LIKE '%MAXVALUE%'
      AND regexp_like(xml_hv, '(19|20)[0-9]{2}')
)
SELECT owner, table_name, ultima_particion, valor_limite, meses_restantes
FROM (
    -- Bloque para tablas RANGE
    SELECT 
        table_owner as owner, 
        table_name, 
        partition_name as ultima_particion,
        CAST(fecha_str AS VARCHAR2(100)) as valor_limite,
        TRIM(TO_CHAR(CASE 
            WHEN fecha_str LIKE '20%' OR fecha_str LIKE '19%' THEN ROUND(MONTHS_BETWEEN(TO_DATE(fecha_str, 'YYYY-MM-DD'), SYSDATE), 1)
            ELSE ROUND(MONTHS_BETWEEN(TO_DATE(SUBSTR(fecha_str, 1, 11), 'DD-MON-YYYY'), SYSDATE), 1)
        END, '999990.9', 'NLS_NUMERIC_CHARACTERS = ''. ''')) as meses_restantes,
        auto_type as auto,
        CASE 
            WHEN fecha_str LIKE '20%' OR fecha_str LIKE '19%' THEN TO_DATE(fecha_str, 'YYYY-MM-DD')
            ELSE TO_DATE(SUBSTR(fecha_str, 1, 11), 'DD-MON-YYYY')
        END as fecha_dt,
        rn
    FROM final_dates
    WHERE rn = 1 AND fecha_str IS NOT NULL
    UNION ALL
    SELECT 
        owner, table_name, ultima_particion, valor_limite,
        TRIM(TO_CHAR(meses_r, '999990.9', 'NLS_NUMERIC_CHARACTERS = ''. ''')) as meses_restantes,
        'NO' as auto,
        fecha_d as fecha_dt,
        rn
    FROM (
        SELECT 
            table_owner as owner, table_name, partition_name as ultima_particion,
            v_limite as valor_limite,
            ROUND(MONTHS_BETWEEN(TO_DATE(v_limite, 'YYYYMM'), SYSDATE), 1) as meses_r,
            TO_DATE(v_limite, 'YYYYMM') as fecha_d,
            ROW_NUMBER() OVER (PARTITION BY table_owner, table_name ORDER BY v_limite DESC) as rn
        FROM (
            SELECT tp.table_owner, tp.table_name, tp.partition_name,
                CAST(REPLACE(EXTRACTVALUE(DBMS_XMLGEN.GETXMLTYPE(
                    'SELECT high_value FROM dba_tab_partitions WHERE table_owner = ''' || tp.table_owner || 
                    ''' AND table_name = ''' || tp.table_name || 
                    ''' AND partition_name = ''' || tp.partition_name || ''''), 
                    '//ROW/HIGH_VALUE'), '''', '') AS VARCHAR2(100)) as v_limite
            FROM dba_tab_partitions tp
            WHERE tp.table_name = 'CM_ORIGINAL_FACTURA'
        )
        WHERE v_limite != 'DEFAULT'
    )
    WHERE rn = 1
)
WHERE fecha_dt > SYSDATE - 90
  AND fecha_dt < ADD_MONTHS(SYSDATE, 12)
  AND auto = 'NO'
ORDER BY fecha_dt ASC;
	exit
	EOF`
								
	if [ -n "$SQLOUTPUT" ]; then
		echo -- "$ORACLE_SID" -- >> "$LOGFILE"
		echo "$SQLOUTPUT" >> "$LOGFILE"
		printf '\n' >> "$LOGFILE"
		QUERIES=YES
	fi
}


#export JAVA_HOME=/usr/java8_64
export HOST=`hostname`
export HOME=/home/oracle
export LOGFILE=${HOME}/partitions_${HOST}.txt
export QUERIES=NO


cd ${HOME}
echo === $HOST > $LOGFILE ==='\n'

ps -ef | grep pmon | awk '/ora_pmon/ {print $0}' | grep -v awk | while read LINE
do
   export ORACLE_SID=`echo ${LINE} | awk '{print $NF}' | awk -F_ '{print $3}'`
   export ORACLE_HOME=`cat /etc/oratab | awk -v pat="$ORACLE_SID" -F ":" 'index(pat, $1) == 1 {print $2}'`
   runQuery $ORACLE_HOME $ORACLE_SID
		
done

if [ ${#QUERIES} -eq 3 ]; then
   scp ${LOGFILE} ususopbd@dwhsrvp01:/ora_scripts/logs/querieslog
   rm ${LOGFILE}
fi

