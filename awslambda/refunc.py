import json
import os
import sys

import six
import kubernetes
from kubernetes import client, config
from kubernetes.client.rest import ApiException
from kubernetes.config.kube_config import KUBE_CONFIG_DEFAULT_LOCATION


def get_funcdef(ns: str, name: str) -> dict:
    rpath = resource_path + "/funcdeves/{name}"
    obj = invoke_api_sync(
        rpath,
        "GET",
        path_params={"namespace": ns, "name": name},
        response_type=object,
        auth_settings=["BearerToken"],
    )
    return obj


def create_funcdef(funcdef: dict):
    """
    create a new refunc
    """

    fndef_obj = funcdef.copy()

    ns = funcdef["metadata"]["namespace"]
    name = funcdef["metadata"].get("name", "")
    if not name:
        name = funcdef["metadata"]["generateName"]

    if "annotations" not in fndef_obj["metadata"]:
        fndef_obj["metadata"]["annotations"] = {}

    rpath = resource_path + "/funcdeves"
    return invoke_api_sync(
        rpath,
        "POST",
        path_params={"namespace": ns},
        header_params={
            "Accept": "application/json",
            "Content-Type": "application/json",
        },
        body=fndef_obj,
        response_type=object,
        auth_settings=["BearerToken"],
    )


def update_funcdef(funcdef: dict):
    """
    update the refunc
    """

    ns, name = funcdef["metadata"]["namespace"], funcdef["metadata"]["name"]

    refunc_obj = get_funcdef(ns, name)
    refunc_obj["spec"] = funcdef["spec"]

    # update annotations
    if "annotations" not in refunc_obj["metadata"]:
        refunc_obj["metadata"]["annotations"] = {}
    refunc_obj["metadata"]["annotations"].update(funcdef["metadata"]["annotations"])

    rpath = resource_path + "/funcdeves/{name}"
    return invoke_api_sync(
        rpath,
        "PUT",
        path_params={"namespace": ns, "name": name},
        header_params={
            "Accept": "application/json",
            "Content-Type": "application/json",
        },
        body=refunc_obj,
        response_type=object,
        auth_settings=["BearerToken"],
    )


def delete_funcdef(ns: str, name: str):
    rpath = resource_path + "/funcdeves/{name}"
    res = invoke_api_sync(
        rpath,
        "DELETE",
        path_params={"namespace": ns, "name": name},
        header_params={
            "Accept": "application/json",
            "Content-Type": "application/json",
        },
        response_type=object,
        auth_settings=["BearerToken"],
    )
    if res["status"] == "Success":
        return {"status": "Success"}

    raise ValueError("Bad status: {}".format(json.dumps(res)))


def list_funcdefs(ns: str, label_selector: str = "") -> [dict]:
    query_params = []
    if label_selector:
        query_params.append(("labelSelector", label_selector))

    rpath = resource_path + "/funcdeves"
    obj = invoke_api_sync(
        rpath,
        "GET",
        path_params={"namespace": ns},
        query_params=query_params,
        response_type=object,
        auth_settings=["BearerToken"],
    )

    return [clean_funcdef_obj(i) for i in obj.get("items", [])]


def ensure_funcdef(funcdef: dict):
    """
    create or update a func
    """
    try:
        return update_funcdef(funcdef)
    except ApiException as e:
        if e.status != 404:
            six.reraise(*sys.exc_info())

    return create_funcdef(funcdef)


def clean_funcdef_obj(obj: dict) -> dict:
    """
    remove extra/private attributes
    """
    obj["metadata"].pop("clusterName", "")
    obj["metadata"].pop("deletionGracePeriodSeconds", "")
    obj["metadata"].pop("deletionTimestamp", "")
    obj["metadata"].pop("generation", "")
    obj["metadata"].pop("resourceVersion", "")
    obj["metadata"].pop("selfLink", "")
    obj["metadata"].pop("uid", "")

    return obj


def invoke_api_sync(
    resource_path,
    method,
    path_params=None,
    query_params=None,
    header_params=None,
    body=None,
    post_params=None,
    files=None,
    response_type=None,
    auth_settings=None,
    _return_http_data_only=None,
    collection_formats=None,
    _preload_content=True,
    _request_timeout=None,
):

    (data, *_) = get_cli().call_api(
        resource_path,
        method,
        path_params,
        query_params,
        header_params,
        body,
        post_params,
        files,
        response_type,
        auth_settings,
        None,
        _return_http_data_only,
        collection_formats,
        _preload_content,
        _request_timeout,
    )
    return data


resource_path = "/apis/k8s.refunc.io/v1beta3/namespaces/{namespace}"

__cli = None


def get_cli() -> kubernetes.client.api_client.ApiClient:
    global __cli
    if not __cli:
        # ensure cfg is loaded
        if os.path.exists(os.path.expanduser(KUBE_CONFIG_DEFAULT_LOCATION)):
            config.load_kube_config()
        else:
            config.incluster_config.load_incluster_config()
        __cli = client.api_client.ApiClient()
    return __cli
