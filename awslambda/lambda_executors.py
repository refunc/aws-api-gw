import os
import re
import json
import time
import logging
import threading
import subprocess

# from datetime import datetime
from multiprocessing import Process, Queue

try:
    from shlex import quote as cmd_quote
except ImportError:
    # for Python 2.7
    from pipes import quote as cmd_quote
from localstack import config
from localstack.utils.common import run, TMP_FILES, to_str

from loguru import logger

# constants
LAMBDA_RUNTIME_PYTHON27 = "python2.7"
LAMBDA_RUNTIME_PYTHON36 = "python3.6"
LAMBDA_RUNTIME_NODEJS = "nodejs"
LAMBDA_RUNTIME_NODEJS610 = "nodejs6.10"
LAMBDA_RUNTIME_NODEJS810 = "nodejs8.10"
LAMBDA_RUNTIME_JAVA8 = "java8"
LAMBDA_RUNTIME_DOTNETCORE2 = "dotnetcore2.0"
LAMBDA_RUNTIME_GOLANG = "go1.x"

LAMBDA_EVENT_FILE = "event_file.json"

# maximum time a pre-allocated container can sit idle before getting killed
MAX_CONTAINER_IDLE_TIME = 540


class LambdaExecutor(object):
    """ Base class for Lambda executors. Subclasses must overwrite the execute method """

    def __init__(self):
        pass

    def execute(
        self,
        func_arn,
        func_details,
        event,
        context=None,
        version=None,
        asynchronous=False,
    ):
        raise Exception("Not implemented.")

    def startup(self):
        pass

    def cleanup(self, arn=None):
        pass

    def run_lambda_executor(self, cmd, env_vars={}, asynchronous=False):
        process = run(
            cmd,
            asynchronous=True,
            stderr=subprocess.PIPE,
            outfile=subprocess.PIPE,
            env_vars=env_vars,
        )
        if asynchronous:
            result = '{"asynchronous": "%s"}' % asynchronous
            log_output = "Lambda executed asynchronously"
        else:
            return_code = process.wait()
            result = to_str(process.stdout.read())
            log_output = to_str(process.stderr.read())

            if return_code != 0:
                raise Exception(
                    "Lambda process returned error status code: %s. Output:\n%s"
                    % (return_code, log_output)
                )
        return result, log_output


class RefuncExecutorLocal(LambdaExecutor):
    def execute(
        self,
        func_arn,
        func_details,
        event,
        context=None,
        version=None,
        asynchronous=False,
    ):
        environment = func_details.envvars.copy()
        lambda_function = func_details.function(version)
        cmd = f"invoke -n {lambda_function['metadata']['namespace']} -t {lambda_function['spec'].get('runtime', {}).get('timeout', MAX_CONTAINER_IDLE_TIME)}s {lambda_function['metadata']['name']}"

        logger.debug(f"cmd: {cmd}")

        process = run(
            cmd,
            asynchronous=True,
            stdin=True,
            stderr=subprocess.PIPE,
            outfile=subprocess.PIPE,
            env_vars=environment,
        )

        process.stdin.write(json.dumps(event).encode("utf-8"))
        process.stdin.flush()
        process.stdin.close()
        return_code = process.wait()
        result = to_str(process.stdout.read())
        log_output = to_str(process.stderr.read())

        if return_code != 0:
            raise Exception(
                "Lambda process returned error status code: %s. Output:\n%s"
                % (return_code, log_output)
            )
        return result, log_output


# --------------
# GLOBAL STATE
# --------------

EXECUTOR_REFUNC = RefuncExecutorLocal()
DEFAULT_EXECUTOR = EXECUTOR_REFUNC
# the keys of AVAILABLE_EXECUTORS map to the LAMBDA_EXECUTOR config variable
AVAILABLE_EXECUTORS = {"refunc": EXECUTOR_REFUNC}
