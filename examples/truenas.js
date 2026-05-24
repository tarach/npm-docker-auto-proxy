const labelContainerName = "subdomain-proxy";

const labels = {
    "npm.proxy.enabled": "true",
    "npm.proxy.domain": "subdomain-proxy.domain.com",
    "npm.proxy.forward_host": "192.168.1.2",
    "npm.proxy.forward_port": "81",
    "npm.proxy.scheme": "http",
    "npm.proxy.websocket": "true",
    "npm.proxy.ssl": "true",
    "npm.proxy.certificate": "*.domain.com",
    "npm.proxy.force_ssl": "true",
    "npm.proxy.http2": "true",
    "npm.proxy.block_exploits": "true",
    "npm.proxy.on_stop": "disable",
};

labelsSetupFunc(labelContainerName, labels)