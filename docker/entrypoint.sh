if [ "${STARLINKS_HOST_NAME}" == '' ]; then 
    echo You should pass HOSTNAME for URL by STARLINKS_HOST_NAME environment variable.
    exit 1
fi

if [ "${STARLINKS_REDIS_DAIL}" == '' ]; then
    echo Redis domain not set. \( Use environment variable STARLINKS_REDIS_DAIL \)
    exit 2;
fi

if [ "${STARLINKS_MYSQL_DSN}" == '' ]; then
    echo Mysql dsn not set. \( Use environment variable STARLINKS_MYSQL_DSN \)
    exit 3;
fi

#if ! [ -f /home/config/nginx.conf ]; then
#    envsubst /home/config/nginx.conf.tmpl >> /home/config/nginx.conf
#fi
sed -E 's/###STARLINKS_HOST_NAME###/'"$STARLINKS_HOST_NAME"'/' /home/config/nginx.conf.tmpl > /etc/nginx/nginx.conf

if [ "$STARLINKS_SECRET" == '' ]; then
    STARLINKS_SECRET=starstudio
fi


nginx  || exit 4


starlinks -api_listen 0.0.0.0:23278 -listen 0.0.0.0:23279 -redis_dail "${STARLINKS_REDIS_DOMAIN}" -secret "${STARLINKS_SECRET}" -sql_dsn "${STARLINKS_MYSQL_DSN}"
