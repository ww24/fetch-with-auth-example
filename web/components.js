const token = "api-test-key";

// maybe doesn't work in Safari
class RestrictedImage extends HTMLImageElement {
    constructor() {
        super();
        const url = this.getAttribute("restricted-src");
        const self = this;
        fetch(url, {
            mode: "cors",
            credentials: "include",
            headers: {
                Authorization: `Bearer ${token}`
            }
        }).then(resp => resp.blob()).then(blob => {
            self.src = URL.createObjectURL(blob);
        }).catch(() => {
            console.error("failed to fetch in RestrictedImage");
        });
    }
}
customElements.define("restricted-image", RestrictedImage, { extends: "img" });

// safari compatible
class RestrictedImageElement extends HTMLElement {
    constructor() {
        super();
        const shadow = this.attachShadow({ mode: "open" });
        const img = new Image();
        const url = this.getAttribute("src");
        const self = this;
        this.getAttributeNames().forEach(key => {
            if (key === "src") return;
            img[key] = self.getAttribute(key);
        });
        fetch(url, {
            mode: "cors",
            credentials: "include",
            headers: {
                Authorization: `Bearer ${token}`
            }
        }).then(resp => resp.blob()).then(blob => {
            img.src = URL.createObjectURL(blob);
        }).catch(() => {
            console.error("failed to fetch in RestrictedImageElement");
        });
        shadow.appendChild(img);
    }
}
customElements.define("restricted-img", RestrictedImageElement);
