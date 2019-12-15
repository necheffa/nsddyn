FROM python:3-alpine3.10
RUN apk update && apk upgrade
COPY *.py /dynupd/
COPY requirements.txt /dynupd/
WORKDIR /dynupd
RUN pip install --trusted-host pypi.python.org -r requirements.txt
CMD ["python3", "main.py"]

