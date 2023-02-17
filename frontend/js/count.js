(function (global, document) {
    'use strict';

    const origin = '[[[.Origin]]]';

    function post(url, data, callback) {
        const xmlDoc = new XMLHttpRequest();

        xmlDoc.open('POST', url, true);
        xmlDoc.setRequestHeader('Content-type', 'application/json');
        xmlDoc.onload = function () {
            callback(JSON.parse(xmlDoc.response));
        };

        xmlDoc.send(JSON.stringify(data));
    }

    let commentsText = function (count) {
        return `${count} ${count === 1 ? 'comment' : 'comments'}`;
    };

    function tags(tag) {
        return document.getElementsByTagName(tag);
    }

    function attrGet(node, a) {
        const attr = node.attributes[a];

        if (attr === undefined) {
            return undefined;
        }

        return attr.value;
    }


    function dataTagsLoad() {
        const scripts = tags('script');
        for (let i = 0; i < scripts.length; i++) {
            if (scripts[i].src.match(/\/js\/count\.js$/)) {
                const customCommentsText = attrGet(scripts[i], 'data-custom-text');
                if (customCommentsText !== undefined) {
                    commentsText = eval(customCommentsText);
                }
            }
        }
    }

    function main() {
        const paths = [];
        const doms = [];
        dataTagsLoad();

        const as = document.getElementsByTagName('a');
        for (let i = 0; i < as.length; i++) {
            let href = as[i].href;
            if (href === undefined) {
                return;
            }

            href = href.replace(/^.*\/\/[^\/]+/, '');

            if (href.endsWith('#comentario')) {
                let pageId = attrGet(as[i], 'data-page-id');
                if (pageId === undefined) {
                    pageId = href.substring(0, href.indexOf('#comentario'));
                    if (pageId.startsWith(parent.location.host)) {
                        pageId = pageId.substring(parent.location.host.length);
                    }
                }

                paths.push(pageId);
                doms.push(as[i]);
            }
        }

        const json = {
            domain: parent.location.host,
            paths: paths,
        };

        post(`${origin}/api/comment/count`, json, function (resp) {
            if (!resp.success) {
                console.log(`[comentario] error: ${resp.message}`);
                return;
            }

            for (let i = 0; i < paths.length; i++) {
                let count = 0;
                if (paths[i] in resp.commentCounts) {
                    count = resp.commentCounts[paths[i]];
                }

                doms[i].innerText = commentsText(count);
            }
        });
    }

    let initted = false;

    function init() {
        if (initted) {
            return;
        }
        initted = true;

        main(undefined);
    }

    const readyLoad = function () {
        const readyState = document.readyState;

        if (readyState === 'loading') {
            // The document is still loading. The div we need to fill might not have
            // been parsed yet, so let's wait and retry when the readyState changes.
            // If there is more than one state change, we aren't affected because we
            // have a double-call protection in init().
            document.addEventListener('readystatechange', readyLoad);
        } else if (readyState === 'interactive') {
            // The document has been parsed and DOM objects are now accessible. While
            // JS, CSS, and images are still loading, we don't need to wait.
            init();
        } else if (readyState === 'complete') {
            // The page has fully loaded (including JS, CSS, and images). From our
            // point of view, this is practically no different from interactive.
            init();
        }
    };

    readyLoad();

}(window, document));
