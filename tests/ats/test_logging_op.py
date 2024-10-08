import logging
from typing import List

import pykube
import pytest
from pytest_helm_charts.clusters import Cluster
from pytest_helm_charts.k8s.deployment import wait_for_deployments_to_run

logger = logging.getLogger(__name__)

namespace_name = "monitoring"
deployment_name= "logging-operator"

timeout: int = 560

@pytest.mark.smoke
def test_api_working(kube_cluster: Cluster) -> None:
    """
    Testing apiserver availability.
    """
    assert kube_cluster.kube_client is not None
    assert len(pykube.Node.objects(kube_cluster.kube_client)) >= 1

# scope "module" means this is run only once, for the first test case requesting! It might be tricky
# if you want to assert this multiple times
# -- Checking that the operator's deployment is present on the cluster
@pytest.fixture(scope="module")
def ic_deployment(kube_cluster: Cluster) -> List[pykube.Deployment]:
    logger.info("Waiting for logging-operator deployment..")

    deployment_ready = wait_for_ic_deployment(kube_cluster)

    logger.info("logging-operator deployment looks satisfied..")

    return deployment_ready

def wait_for_ic_deployment(kube_cluster: Cluster) -> List[pykube.Deployment]:
    deployments = wait_for_deployments_to_run(
        kube_cluster.kube_client,
        [deployment_name],
        namespace_name,
        timeout,
    )
    return deployments


@pytest.fixture(scope="module")
def pods(kube_cluster: Cluster) -> List[pykube.Pod]:
    pods = pykube.Pod.objects(kube_cluster.kube_client)

    pods = pods.filter(namespace=namespace_name, selector={
                       'app.kubernetes.io/name': 'logging-operator', 'app.kubernetes.io/instance': 'logging-operator'})

    return pods

# when we start the tests on circleci, we have to wait for pods to be available, hence
# this additional delay and retries
# -- Checking that all pods from the operator's deployment are available (i.e in "Ready" state)
@pytest.mark.smoke
@pytest.mark.upgrade
@pytest.mark.flaky(reruns=5, reruns_delay=10)
def test_pods_available(ic_deployment: List[pykube.Deployment]):
    for s in ic_deployment:
        assert int(s.obj["status"]["readyReplicas"]) == int(
            s.obj["spec"]["replicas"])
