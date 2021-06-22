#!/bin/bash

# имена сервисов
service[0]="./access"

# пути запуска
service_path[0]="/home/admin/go/src/access"
#
while true
do
    for index in ${!service[*]}
    do
        # проверяем процесс
        process_run=`ps ax | grep -v grep | grep ${service[$index]}`
        # если не запущен, тогда стартуем
        if [ "${process_run:-null}" = null ]; then
            # переход в директорию
            cd "${service_path[$index]}"
            # го!
            `${service[$index]} > 0.txt &`
            sleep 3
        fi
    done
    sleep 1
done