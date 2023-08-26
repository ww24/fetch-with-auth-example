class TokenStore {
    #promise;
    #resolve;
    #isResolved = false;

    constructor() {
        const self = this;
        this.#promise = new Promise(resolve => {
            self.#resolve = resolve;
        }).then(v => {
            self.#isResolved = true;
            return v;
        });
    }

    #timeout(timeout) {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                reject("rejected because timed out");
            }, timeout);
        });
    }

    getToken(timeout = 0) {
        if (timeout === 0) {
            return this.#promise;
        }
        return Promise.race([this.#promise, this.#timeout(timeout)]);
    }

    set token(token) {
        if (this.#isResolved) {
            this.#promise = Promise.resolve(token);
            return;
        }
        this.#resolve(token);
    }
}
const store = new TokenStore();

self.addEventListener("install", event => {
    event.waitUntil(self.skipWaiting());
});

self.addEventListener("activate", event => {
    event.waitUntil(self.clients.claim());
});

const interceptor = async (request, token) => {
    const headers = new Headers(request.headers);
    headers.set("Authorization", `Bearer ${await token}`);
    const req = new Request(request, {
        mode: "cors", // cannot rewrite header in no-cors mode
        headers
    });

    const resp = await fetch(req);

    // debug
    // showHeaders("original:", request.headers);
    // showHeaders("new:", req.headers);
    // console.log(req);
    // console.log(resp);

    return resp;
};

self.addEventListener("fetch", event => {
    const req = event.request;
    if (
        req.method !== "GET" ||
        req.headers.get("Authorization") != null ||
        !req.url.includes("/restricted/") ||
        req.url.includes("#skip-service-worker")) {
        // do nothing
        return;
    }
    event.respondWith(interceptor(event.request, store.getToken(3000)));
});

self.addEventListener("message", event => {
    if (event.data == null) return;
    switch (event.data.action) {
        case "sync-token":
            store.token = event.data.token;
            return;
    }
});

function showHeaders(prefix, headers) {
    for (const key of headers.keys()) {
        console.log(prefix, key + ":", headers.get(key));
    }
}
