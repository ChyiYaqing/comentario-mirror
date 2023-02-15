export class HttpClientError {
    constructor(
        readonly status: number,
        readonly message: string,
        readonly response: any,
    ) {}
}

export class HttpClient {

    constructor(
        readonly baseUrl: string,
    ) {}

    /**
     * Run an HTTP POST request to the given endpoint.
     * @param path Endpoint path, relative to the client's baseURl.
     * @param body Optional request body.
     */
    post<T>(path: string, body?: any): Promise<T> {
        return this.request<T>('POST', path, body);
    }

    /**
     * Run an HTTP GET request to the given endpoint.
     * @param path Endpoint path, relative to the client's baseURl.
     */
    get<T>(path: string): Promise<T> {
        return this.request<T>('GET', path);
    }

    /**
     * Convert the relative endpoint path to an absolute one by prepending it with the base URL.
     * @param path Relative endpoint path.
     * @private
     */
    private getEndpointUrl(path: string): string {
        // Combine the two paths, making sure there's exactly one slash in between
        return this.baseUrl + (this.baseUrl.endsWith('/') ? '' : '/') + (path.startsWith('/') ? path.substring(1) : path);
    }

    private request<T>(method: 'POST' | 'GET', path: string, body?: any): Promise<T> {
        return new Promise((resolve, reject) => {
            try {
                // Prepare an XMLHttpRequest
                const req = new XMLHttpRequest();
                req.open(method, this.getEndpointUrl(path), true);
                if (body) {
                    req.setRequestHeader('Content-type', 'application/json');
                }

                // Resolve or reject the promise on load, based on the return status
                const handleError = () => reject(new HttpClientError(req.status, req.statusText, req.response));
                req.onload = () => {
                    if (req.status >= 200 && req.status <= 299) {
                        resolve(JSON.parse(req.response));
                    } else {
                        handleError();
                    }
                };
                req.onerror = handleError;

                // Run the request
                req.send(body ? JSON.stringify(body) : undefined);

            } catch (e) {
                // Reject the promise on any failure
                reject(e);
            }
        });
    }
}
