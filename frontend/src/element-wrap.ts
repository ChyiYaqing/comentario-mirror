/**
 * Wrapper around an HTML element that facilitates adjusting and managing it.
 */
export class Wrap<T extends HTMLElement> {

    static readonly idPrefix = 'comentario-';

    constructor(
        private el: T,
    ) {}

    /**
     * Instantiate a new element with the given tag name, and return a new Wrap object for it.
     * @param tagName Name of the tag to create an element with.
     */
    static new<K extends keyof HTMLElementTagNameMap>(tagName: K): Wrap<HTMLElementTagNameMap[K]> {
        return new Wrap(document.createElement(tagName));
    }

    /**
     * Find an existing element with the given ID (optionally prepending it with idPrefix). Whether the element actually
     * exists, can be derived from the ok property.
     * @param id ID of the element to find (excluding the prefix).
     * @param noPrefix Whether skip prepending the ID with idPrefix.
     */
    static byId<K extends keyof HTMLElementTagNameMap>(id: string, noPrefix?: boolean): Wrap<HTMLElementTagNameMap[K]> {
        return new Wrap(document.getElementById(noPrefix ? id : this.idPrefix + id) as HTMLElementTagNameMap[K]);
    }

    /**
     * Whether the underlying element is present.
     */
    get ok(): boolean {
        return !!this.el;
    }

    /**
     * Set attributes of the underlying element from the provided object.
     * @param values Object that provides attribute names (keys) and their values. null and undefined values cause
     * attribute removal from the node.
     */
    attr(values: { [k: string]: string }): Wrap<T> {
        if (this.el) {
            Object.keys(values).forEach(k => {
                const v = values[k];
                if (v === undefined || v === null) {
                    this.el.removeAttribute(k);
                } else {
                    this.el.setAttribute(k, v);
                }
            });
        }
        return this;
    }

    /**
     * Set the ID of the underlying element;
     * @param s New value to set.     */
    id(s: string): Wrap<T> {
        if (this.el) {
            this.el.id = s;
        }
        return this;
    }

    /**
     * Set the innerText of the underlying element;
     * @param s New value to set.
     */
    inner(s: string): Wrap<T> {
        if (this.el) {
            this.el.innerText = s;
        }
        return this;
    }

    /**
     * Set the innerHTML of the underlying element;
     * @param s New value to set.
     */
    html(s: string): Wrap<T> {
        if (this.el) {
            this.el.innerHTML = s;
        }
        return this;
    }

    /**
     * Set the style of the underlying element;
     * @param s New value to set.
     */
    style(s: string): Wrap<T> {
        return this.attr({style: s});
    }

    /**
     * Insert the underlying element as the first child to the specified parent.
     * @param parent Wrapper of the new parent for the element.     */
    prependTo(parent: Wrap<any>): Wrap<T> {
        if (this.el && parent.el) {
            parent.el.prepend(this.el);
        }
        return this;
    }

    /**
     * Append the underlying element as the last child to the specified parent.
     * @param parent Wrapper of the new parent for the element.
     */
    appendTo(parent: Wrap<any>): Wrap<T> {
        if (this.el && parent.el) {
            parent.el.appendChild(this.el);
        }
        return this;
    }

    /**
     * Append the specified elements as children to the underlying element.
     * @param children Wrapped child elements to add.
     */
    append(...children: Wrap<any>[]): Wrap<T> {
        if (this.el) {
            children.forEach(w => w.ok && this.el.appendChild(w.el));
        }
        return this;
    }

    /**
     * Remove the underlying element from the DOM.
     */
    remove(): Wrap<T> {
        this.el?.parentNode.removeChild(this.el);
        return this;
    }

    /**
     * Insert the underlying element as the next sibling to the given element.
     * @param sibling Wrapper of new sibling for the element.
     */
    insertAfter(sibling: Wrap<any>): Wrap<T> {
        if (this.el && sibling.el) {
            sibling.el.parentNode.insertBefore(this.el, sibling.el.nextSibling);
        }
        return this;
    }

    /**
     * Add the specified classes to the underlying element.
     * @param classes Class(es) to add. Falsy values are ignored.
     */
    classes(...classes: string[]): Wrap<T> {
        if (this.el) {
            classes?.forEach(c => c && this.el.classList.add(`commento-${c}`));
        }
        return this;
    }

    /**
     * Remove the provided class or classes from the underlying element.
     * @param classes Class(es) to remove. Falsy values are ignored.
     */
    noClasses(...classes: string[]): Wrap<T> {
        if (this.el) {
            classes.forEach(c => c && this.el.classList.remove(`commento-${c}`));
        }
        return this;
    }

    /**
     * Bind a handler to the onClick event of the underlying element.
     * @param handler Handler to bind.
     */
    click(handler: () => void): Wrap<T> {
        this.el?.addEventListener('click', handler);
        return this;
    }

    /**
     * Bind a handler to the onLoad event of the underlying element.
     * @param handler Handler to bind.
     */
    load(handler: () => void): Wrap<T> {
        this.el?.addEventListener('load', handler);
        return this;
    }

    /**
     * Remove all event listeners from the underlying element.
     * NB: This method can cause a replacement of the underlying element.
     */
    unlisten(): Wrap<T> {
        if (this.el) {
            const clone = this.el.cloneNode(true) as T;
            this.el.parentNode?.replaceChild(clone, this.el);
            this.el = clone;
        }
        return this;
    }

    /**
     * Set the value of the checked property of the underlying (input) element.
     * @param b New value of checked.
     */
    checked(b: boolean): Wrap<T> {
        if (this.el) {
            (this.el as unknown as HTMLInputElement).checked = b;
        }
        return this;
    }

    /**
     * Scroll to the underlying element.
     */
    scrollTo(): Wrap<T> {
        this.el?.scrollIntoView({block: 'start', inline: 'nearest', behavior: 'smooth'});
        return this;
    }

    /**
     * Enables automatic height adjusting of the underlying textarea.
     */
    autoExpand(): Wrap<T> {
        this.el.addEventListener('input', evt => {
            (evt.target as HTMLTextAreaElement).style.height = '';
            const h = Math.min(Math.max((evt.target as HTMLTextAreaElement).scrollHeight + 16, 75), 400);
            (evt.target as HTMLTextAreaElement).style.height = `${h}px`;
        });
        return this;
    }

    /**
     * Return a value of the attribute of the underlying element with the given name.
     * @param attrName Attribute name.
     */
    getAttr(attrName: string): string {
        const attr = this.el?.attributes.getNamedItem(attrName);
        return attr === undefined ? undefined : attr?.value;
    }
}
