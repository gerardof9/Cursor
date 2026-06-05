#!/bin/sh

export JAVA_HOME=/usr/java8_64

DBSAT_HOME=/home/oracle/dbsat
CONF=${DBSAT_HOME}/Discover/conf
ORATAB=/etc/oratab

DBSAT_USER=batch
DBSAT_PASS='BATCH'

DEFAULT_PORT=1521

REMOTE_USER=ususopbd
REMOTE_HOST=dwhsrvp01
REMOTE_DIR=/ora_scripts/logs/dbsat

export DBSAT_HOME
export CONF
export ORATAB

cd "${DBSAT_HOME}" || exit 1

log()
{
    echo "`date '+%Y-%m-%d %H:%M:%S'` - $*"
}

run_sql_sysdba()
{
sqlplus -s "/ as sysdba" <<EOF
set pages 0
set heading off
set feedback off
set verify off
set trimspool on
set lines 200
whenever sqlerror exit failure
$1
exit
EOF
}

get_oracle_home()
{
    awk -F: -v sid="$1" '
    $1 == sid {
        print $2
        exit
    }' "${ORATAB}"
}

is_cdb()
{
    RESULT=`run_sql_sysdba "select cdb from v\\$database;" 2>/dev/null`

    echo "${RESULT}" | grep -q "^YES$"
}

get_pdbs()
{
run_sql_sysdba "
select name
from v\\$containers
where name not in ('CDB\\$ROOT','PDB\\$SEED')
and open_mode='READ WRITE';
"
}

generate_discover_config()
{
    TARGET_SERVICE="$1"
    CONFIG_FILE="$2"

    HOSTNAME_VALUE=`hostname`

    log "Generando ${CONFIG_FILE}"


    sed "s/DB_HOSTNAME =.*/DB_HOSTNAME = `hostname`/" \
        "${CONF}/sample_dbsat.config" > "${CONF}/tmp.$$"

    sed "s/DB_SERVICE_NAME =.*/DB_SERVICE_NAME = ${TARGET_SERVICE}/" \
        "${CONF}/tmp.$$" > "${CONF}/dbsat_${TARGET_SERVICE}.config"

    rm -f "${CONF}/tmp.$$"

    log "Archivo generado correctamente: ${CONFIG_FILE}"

    return 0
}

run_dbsat()
{
    TARGET_SERVICE="$1"

    CONFIG_FILE="${CONF}/dbsat_${TARGET_SERVICE}.config"

    generate_discover_config "${TARGET_SERVICE}" "${CONFIG_FILE}"

    if [ $? -ne 0 ]
    then
        log "ERROR: no fue posible generar ${CONFIG_FILE}"
        return 1
    fi

    log "Inicio DBSAT para ${TARGET_SERVICE}"

    ${DBSAT_HOME}/dbsat collect -n "/ as sysdba" dbsat_${TARGET_SERVICE}

    if [ $? -ne 0 ]
    then
        log "ERROR: dbsat collect falló para ${TARGET_SERVICE}"
        return 1
    fi

    ${DBSAT_HOME}/dbsat report -n dbsat_${TARGET_SERVICE}

    if [ $? -ne 0 ]
    then
        log "ERROR: dbsat report falló para ${TARGET_SERVICE}"
        return 1
    fi

    echo "${DBSAT_USER}\n${DBSAT_PASS}\n" | ${DBSAT_HOME}/dbsat discover -n -c "${CONFIG_FILE}" dbsat_${TARGET_SERVICE}

    if [ $? -ne 0 ]
    then
        log "ERROR: dbsat discover falló para ${TARGET_SERVICE}"
        return 1
    fi

    return 0
}

send_results()
{
    TARGET_DIR="$1"

    scp -r "${DBSAT_HOME}/${TARGET_DIR}" ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}

    if [ $? -ne 0 ]
    then
        log "WARNING: SCP falló para ${TARGET_DIR}"
    fi
}

process_target()
{
    TARGET_SERVICE="$1"

    rm -rf "${TARGET_SERVICE}"

    mkdir "${TARGET_SERVICE}"

    if [ $? -ne 0 ]
    then
        log "ERROR: no se pudo crear directorio ${TARGET_SERVICE}"
        return 1
    fi

    cd "${TARGET_SERVICE}" || return 1

    run_dbsat "${TARGET_SERVICE}"
    RC=$?

    cd "${DBSAT_HOME}" || return 1

    if [ $RC -eq 0 ]
    then
        send_results "${TARGET_SERVICE}"
    else
        log "ERROR: procesamiento fallido para ${TARGET_SERVICE}. No se realiza SCP."
    fi

    log "Fin procesamiento ${TARGET_SERVICE}"

    return $RC
}

ps -ef | awk '$NF ~ /^ora_pmon_/ {print $NF}' |
while read PMON
do
    INSTANCE_SID=`echo "${PMON}" | awk -F_ '{print $3}'`

    export ORACLE_SID="${INSTANCE_SID}"

    unset ORACLE_PDB_SID

    ORACLE_HOME=`get_oracle_home "${INSTANCE_SID}"`

    if [ -z "${ORACLE_HOME}" ]
    then
        log "ERROR: ORACLE_HOME no encontrado para ${INSTANCE_SID}"
        continue
    fi

    export ORACLE_HOME
    export PATH=${ORACLE_HOME}/bin:${PATH}

    log "Procesando instancia ${INSTANCE_SID}"

    if is_cdb
    then
        log "${INSTANCE_SID} identificada como CDB"

        for PDB in `get_pdbs`
        do
            export ORACLE_PDB_SID="${PDB}"

            log "Procesando PDB ${PDB}"

            process_target "${PDB}"

            unset ORACLE_PDB_SID
        done
    else
        log "${INSTANCE_SID} identificada como NON-CDB"

        unset ORACLE_PDB_SID

        process_target "${INSTANCE_SID}"
    fi
done

log "Proceso DBSAT finalizado"