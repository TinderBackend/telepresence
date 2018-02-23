import os
from sys import executable
from json import (
    JSONDecodeError,
    loads, dumps,
)
from shutil import which
from subprocess import (
    CalledProcessError,
    PIPE, STDOUT,
    Popen,
    check_output, check_call,
)

from pathlib import Path

from telepresence.utilities import find_free_port
from .utils import (
    KUBECTL,
    DIRECTORY,
    random_name,
    run_webserver,
    create_namespace,
    telepresence_version,
)

REGISTRY = os.environ.get("TELEPRESENCE_REGISTRY", "datawire")


class _ContainerMethod(object):
    name = "container"

    def unsupported(self):
        missing = set()
        for exe in {"socat", "docker"}:
            if which(exe) is None:
                missing.add(exe)
        if missing:
            return "Required executables {} not found on $PATH".format(
                missing,
            )
        return None


    def command_has_graceful_failure(self, command):
        return False


    def loopback_is_host(self):
        return False


    def inherits_client_environment(self):
        return False


    def telepresence_args(self, probe):
        return [
            "--method", "container",
            "--docker-run",
            "--volume", "{}:/probe.py".format(probe),
            "python:3-alpine",
            "python", "/probe.py",
        ]


class _InjectTCPMethod(object):
    name = "inject-tcp"

    def unsupported(self):
        return None


    def command_has_graceful_failure(self, command):
        return command in {
            "ping", "traceroute", "nslookup", "host", "dig",
        }


    def loopback_is_host(self):
        return True


    def inherits_client_environment(self):
        return True


    def telepresence_args(self, probe):
        return [
            "--method", "inject-tcp",
            "--run", executable, probe,
        ]



class _VPNTCPMethod(object):
    name = "vpn-tcp"

    def unsupported(self):
        return None


    def command_has_graceful_failure(self, command):
        return command in {
            "ping", "traceroute",
        }


    def loopback_is_host(self):
        return True


    def inherits_client_environment(self):
        return True


    def telepresence_args(self, probe):
        return [
            "--method", "vpn-tcp",
            "--run", executable, probe,
        ]



class _ExistingDeploymentOperation(object):
    def __init__(self, swap):
        self.swap = swap
        if swap:
            self.name = "swap"
        else:
            self.name = "existing"


    def inherits_deployment_environment(self):
        return True


    def prepare_deployment(self, deployment_ident, environ):
        if self.swap:
            image = "openshift/hello-openshift"
        else:
            image = "{}/telepresence-k8s:{}".format(
                REGISTRY,
                telepresence_version(),
            )
        create_deployment(deployment_ident, image, environ)


    def telepresence_args(self, deployment_ident):
        if self.swap:
            option = "--swap-deployment"
        else:
            option = "--deployment"
        return [
            "--namespace", deployment_ident.namespace,
            option, deployment_ident.name,
        ]



class _NewDeploymentOperation(object):
    name = "new"

    def inherits_deployment_environment(self):
        return False


    def prepare_deployment(self, deployment_ident, environ):
        pass


    def telepresence_args(self, deployment_ident):
        return [
            "--namespace", deployment_ident.namespace,
            "--new-deployment", deployment_ident.name,
        ]



def create_deployment(deployment_ident, image, environ):
    def env_arguments(environ):
        return list(
            "--env={}={}".format(k, v)
            for (k, v)
            in environ.items()
        )
    deployment = dumps({
        "kind": "Deployment",
        "apiVersion": "extensions/v1beta1",
        "metadata": {
            "name": deployment_ident.name,
            "namespace": deployment_ident.namespace,
        },
        "spec": {
            "replicas": 2,
            "template": {
                "metadata": {
                    "labels": {
                        "name": deployment_ident.name,
                        "telepresence-test": deployment_ident.name,
                        "hello": "monkeys",
                    },
                },
                "spec": {
                    "volumes": [{
                        "name": "podinfo",
                        "downwardAPI": {
                            "items": [{
                                "path": "labels",
                                "fieldRef": {"fieldPath": "metadata.labels"},
                            }],
                        },
                    }],
                    "containers": [{
                        "name": "hello",
                        "image": image,
                        "env": list(
                            {"name": k, "value": v}
                            for (k, v)
                            in environ.items()
                        ),
                        "volumeMounts": [{
                            "name": "podinfo",
                            "mountPath": "/podinfo",
                        }],
                    }],
                },
            },
        },
    })
    check_output([KUBECTL, "create", "-f", "-"], input=deployment.encode("utf-8"))


METHODS = [
    _ContainerMethod(),
    _InjectTCPMethod(),
    _VPNTCPMethod(),
]
OPERATIONS = [
    _ExistingDeploymentOperation(False),
    _ExistingDeploymentOperation(True),
    _NewDeploymentOperation(),
]

class ResourceIdent(object):
    """
    Identify a Kubernetes resource.
    """
    def __init__(self, namespace, name):
        self.namespace = namespace
        self.name = name


def _cleanup_deployment(ident):
    check_call([
        KUBECTL, "delete",
        "--namespace", ident.namespace,
        "--ignore-not-found",
        "deployment", ident.name,
    ])



def _telepresence(telepresence_args, env=None):
    """
    Run a probe in a Telepresence execution context.

    :param list telepresence: Arguments to pass to the Telepresence CLI.

    :param env: Environment variables to set for the Telepresence CLI.  These
        are added to the current process's environment.  ``None`` means the
        same thing as ``{}``.
    """
    args = [
        executable,
        which("telepresence"),
        "--logfile=-",
    ] + telepresence_args

    pass_env = os.environ.copy()
    if env is not None:
        pass_env.update(env)

    print("Running {}".format(args))
    return check_output(
        args=args,
        stdin=PIPE,
        stderr=STDOUT,
        env=pass_env,
    )



def run_telepresence_probe(
        request,
        method,
        operation,
        desired_environment,
        client_environment,
        probe_urls,
        probe_commands,
        probe_paths,
):
    """
    :param request: The pytest mumble mumble whatever.

    :param method: The definition of a Telepresence method to use for this
        run.

    :param operation: The definition of a Telepresence operation to use
        for this run.

    :param dict desired_environment: Key/value pairs to set in the probe's
        environment.

    :param dict client_environment: Key/value pairs to set in the Telepresence
        CLI's environment.

    :param list[str] probe_urls: URLs to direct the probe process to request
        and return to us.

    :param list[str] probe_commands: Commands (argv[0]) to direct the probe to
        attempt to run and report back on the results.

    :param list[str] probe_paths: Paths relative to $TELEPRESENCE_ROOT to
        direct the probe to read and report back to us.
    """
    probe_endtoend = (Path(__file__).parent / "probe_endtoend.py").as_posix()

    # Create a web server service.  We'll observe side-effects related to
    # this, such as things set in our environment, and also interact with
    # it directly to demonstrate behaviors related to networking.  It's
    # important that we create this before the ``prepare_deployment`` step
    # below because the environment supplied by Kubernetes to a
    # Deployment's containers depends on the state of the cluster at the
    # time of pod creation.
    deployment_ident = ResourceIdent(
        namespace=random_name(),
        name=random_name(),
    )
    create_namespace(deployment_ident.namespace, deployment_ident.name)

    # TODO: Factor run_webserver into a fixture that Probe can manage so that
    # run_telepresence_probe can just focus on running telepresence.

    # This is an extra pod running on Kubernetes so that various tests can
    # observe how such a thing impacts on the Telepresence execution
    # environment (e.g., environment variables set, etc).
    webserver_name = run_webserver(deployment_ident.namespace)

    operation.prepare_deployment(deployment_ident, desired_environment)
    print("Prepared deployment {}/{}".format(deployment_ident.namespace, deployment_ident.name))
    request.addfinalizer(lambda: _cleanup_deployment(deployment_ident))

    probe_args = []
    for url in probe_urls:
        probe_args.extend(["--probe-url", url])
    for command in probe_commands:
        probe_args.extend(["--probe-command", command])
    for path in probe_paths:
        probe_args.extend(["--probe-path", path])

    operation_args = operation.telepresence_args(deployment_ident)
    method_args = method.telepresence_args(probe_endtoend)
    args = operation_args + method_args + probe_args
    try:
        output = _telepresence(args, client_environment)
    except CalledProcessError as e:
        assert False, "Failure running {}: {}\n{}".format(
            ["telepresence"] + args, str(e), e.output.decode("utf-8"),
        )
    else:
        # Scrape the payload out of the overall noise.
        setup_logs, output, teardown_logs = output.decode("utf-8").split(
            u"{probe delimiter}",
        )
        try:
            probe_result = loads(output)
        except JSONDecodeError:
            assert False, "Could not decode JSON probe result from {}:\n{}".format(
                ["telepresence"] + args, output.decode("utf-8"),
            )
        print("Telepresence output:\n{}{}".format(setup_logs, teardown_logs))
        return ProbeResult(webserver_name, probe_result)



class ProbeResult(object):
    def __init__(self, webserver_name, result):
        self.webserver_name = webserver_name
        self.result = result



class Probe(object):
    CLIENT_ENV_VAR = "SHOULD_NOT_BE_SET"

    DESIRED_ENVIRONMENT = {
        "MYENV": "hello",
        "EXAMPLE_ENVFROM": "foobar",

        # XXXX
        # Container method doesn't support multi-line environment variables.
        # Therefore disable this for all methods or the container tests all
        # fail...
        # XXXX
        # "EX_MULTI_LINE": (
        #     "first line (no newline before, newline after)\n"
        #     "second line (newline before and after)\n"
        # ),
    }

    # A resource available from a server running on the Telepresence host
    # which the tests can use to verify correct routing-to-host behavior from
    # the Telepresence execution context.
    LOOPBACK_URL_TEMPLATE = "http://localhost:{}/test_endtoend.py"

    # Commands which indirectly interact with Telepresence in some way and
    # which may not be s upported (and which we care about them failing in a
    # nice way).
    QUESTIONABLE_COMMANDS = [
        "ping",
        "traceroute",
        "nslookup",
        "host",
        "dig",
    ]

    # Paths relative to $TELEPRESENCE_ROOT in the Telepresence execution
    # context which the probe will read and return to us.
    INTERESTING_PATHS = [
        "podinfo/labels",
        "var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
    ]

    _result = None

    def __init__(self, request, method, operation):
        self._request = request
        self.method = method
        self.operation = operation
        self._cleanup = []

    def __str__(self):
        return "Probe[{}, {}]".format(
            self.method.name,
            self.operation.name,
        )


    def result(self):
        if self._result is None:
            print("Launching {}".format(self))

            local_port = find_free_port()
            self.loopback_url = self.LOOPBACK_URL_TEMPLATE.format(
                local_port,
            )
            # This is a local web server that the Telepresence probe can try to
            # interact with to verify network routing to the host.
            p = Popen(
                # TODO Just cross our fingers and hope this port is available...
                [executable, "-m", "http.server", str(local_port)],
                cwd=str(DIRECTORY),
            )
            self._cleanup.append(lambda: [p.terminate(), p.wait()])

            self._result = run_telepresence_probe(
                self._request,
                self.method,
                self.operation,
                self.DESIRED_ENVIRONMENT,
                {self.CLIENT_ENV_VAR: "FOO"},
                [self.loopback_url],
                self.QUESTIONABLE_COMMANDS,
                self.INTERESTING_PATHS,
            )
        return self._result


    def cleanup(self):
        print("Cleaning up {}".format(self))
        for cleanup in self._cleanup:
            cleanup()
