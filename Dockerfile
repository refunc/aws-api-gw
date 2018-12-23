FROM python:3.7

COPY requirements.txt /opt/aws-api-gw/

WORKDIR /opt/aws-api-gw

RUN export PIP_DOWNLOAD_CACHE=/tmp \
    && pip install -r requirements.txt  \
    && rm -rf /tmp/* && rm -rf ~/.cache
