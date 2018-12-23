import os
from loguru import logger


if "S3_ENDPOINT" in os.environ:
    from localstack import constants

    # update region
    region = os.environ.get("S3_REGION", "us-east-1")
    region = region if region else constants.REGION_LOCAL
    constants.REGION_LOCAL = region
    logger.debug(f"region: {region}")

    from localstack.utils.aws import aws_stack
    import boto3

    s3_endpoint = os.environ["S3_ENDPOINT"]
    logger.debug(f"s3_endponit: {s3_endpoint}")

    orgin_get_local_service_url = aws_stack.get_local_service_url

    def get_local_service_url(service_name):
        if service_name == "s3api":
            service_name = "s3"
        if service_name == "s3":
            return s3_endpoint
        return orgin_get_local_service_url(service_name)

    aws_stack.get_local_service_url = get_local_service_url

    aws_stack.CUSTOM_BOTO3_SESSION = boto3.Session(
        aws_access_key_id=os.environ.get("AWS_ACCESS_KEY_ID", ""),
        aws_secret_access_key=os.environ.get("AWS_SECRET_ACCESS_KEY", ""),
        region_name=region,
    )
    os.environ["ENV"] = "refunc"
    aws_stack.PREDEFINED_ENVIRONMENTS.update(
        {"refunc": aws_stack.Environment(region, "refunc")}
    )

from awslambda import lambda_api
from localstack.utils.common import FuncThread, TMP_THREADS

PORT_LAMBDA = int(os.environ.get("PORT_LAMBDA", "4574"))


def start_lambda(port=PORT_LAMBDA, asynchronous=False):
    return start_local_api(
        "Lambda", port, method=lambda_api.serve, asynchronous=asynchronous
    )


def start_local_api(name, port, method, asynchronous=False):
    print("Starting mock %s service (http port %s)..." % (name, port))
    if asynchronous:
        thread = FuncThread(method, port, quiet=True)
        thread.start()
        TMP_THREADS.append(thread)
        return thread
    else:
        method(port)


if __name__ == "__main__":
    start_lambda()
