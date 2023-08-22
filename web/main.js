function waitServiceWorkerActivated() {
    return new Promise((resolve) => {
        const onStatechange = (event) => {
            if (navigator.serviceWorker.controller.state === "activated") {
                if (event != null) {
                    navigator.serviceWorker.controller.removeEventListener("statechange", onStatechange);
                }
                resolve(navigator.serviceWorker.controller);
                return true;
            }
            return false;
        };
        const onControllerchange = (event) => {
            if (navigator.serviceWorker.controller) {
                if (event != null) {
                    navigator.serviceWorker.removeEventListener("controllerchange", onControllerchange);
                }
                if (onStatechange()) {
                    return true;
                }
                navigator.serviceWorker.controller.addEventListener("statechange", onStatechange);
                return true;
            }
            return false;
        };
        if (onControllerchange()) {
            return;
        }
        navigator.serviceWorker.addEventListener("controllerchange", onControllerchange);
    });
}

function retry(target) {
    console.log("retrying...");
    const retryFlag = "#retry";
    if (target.src.includes(retryFlag)) return;
    waitServiceWorkerActivated().then(() => {
        target.src += retryFlag;
    });
}

(() => {
    const token = "api-test-key";

    if (navigator.serviceWorker == null) {
        console.error("ServiceWorker is not supported");
    } else {
        navigator.serviceWorker.register("sw.js", {
            scope: "/",
        }).then(waitServiceWorkerActivated).then(active => {
            console.log("ServiceWorker is activated");
            active.postMessage({ action: "sync-token", token });
        }).catch(err => {
            console.error(`ServiceWorker registration failed with ${err}`);
        });
    }
})();
