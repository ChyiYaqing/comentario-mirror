/**
 * Wrapper around an HTML element that facilitates adjusting and managing it.
 */
export class Wrap<T extends HTMLElement> {

    static readonly idPrefix = 'comentario-';

    constructor(
        private el?: T,
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
     * The underlying HTML element. Throws an error if no element present.
     */
    get element(): T {
        if (!this.el) {
            throw new Error('No underlying HTML element in the Wrap');
        }
        return this.el;
    }

    /**
     * Whether the underlying element is present.
     */
    get ok(): boolean {
        return !!this.el;
    }

    /**
     * Whether the underlying element has children.
     */
    get hasChildren(): boolean {
        return !!this.el?.childNodes?.length;
    }

    /**
     * Inner text of the underlying element.
     */
    get innerText(): string {
        return this.el?.innerText;
    }

    /**
     * Value of the underlying element.
     */
    get val(): string {
        return (this.el as any)?.value;
    }

    /**
     * Whether the underlying (input) element is checked.
     */
    get isChecked(): boolean {
        return (this.el as unknown as HTMLInputElement)?.checked;
    }

    /**
     * Set attributes of the underlying element from the provided object.
     * @param values Object that provides attribute names (keys, they can use camelCase, which will be converted to
     * kebab-case) and their values. null and undefined values cause attribute removal from the node.
     */
    attr(values: { [k: string]: string }): Wrap<T> {
        if (this.el && values) {
            Object.keys(values).forEach(k => {
                const v = values[k];
                // Convert the cameCase attribute name into kebab-case
                k = k.replace(/[A-Z]/g, l => `-${l.toLowerCase()}`);
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
            this.el.id = Wrap.idPrefix + s;
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
     * Set the value of the underlying element;
     * @param s New value to set.
     */
    value(s: string): Wrap<T> {
        if (this.el) {
            (this.el as any).value = s;
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
     * @param parent Wrapper of the new parent for the element.
     */
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
     * @param children Wrapped child elements to add. Falsy and empty wrappers are skipped.
     */
    append(...children: Wrap<any>[]): Wrap<T> {
        if (this.el) {
            children.forEach(w => w?.ok && this.el.appendChild(w.el));
        }
        return this;
    }

    /**
     * Replace the underlying element in the DOM with the given one.
     */
    replaceWith(newEl: Wrap<any>): Wrap<T> {
        if (newEl.el) {
            this.el?.replaceWith(newEl.el);
        }
        return this;
    }

    /**
     * Remove the underlying element from the DOM.
     */
    remove(): Wrap<T> {
        this.el?.parentNode?.removeChild(this.el);
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
    click(handler: (target: Wrap<T>, e: MouseEvent) => void): Wrap<T> {
        this.el?.addEventListener('click', e => handler(this, e));
        return this;
    }

    /**
     * Bind a handler to the onKeydown event of the underlying element.
     * @param handler Handler to bind.
     */
    keydown(handler: (target: Wrap<T>, e: KeyboardEvent) => void): Wrap<T> {
        this.el?.addEventListener('keydown', e => handler(this, e));
        return this;
    }

    /**
     * Bind a handler to the given event of the underlying element.
     * @param type Event type to bind the handler to.
     * @param handler Handler to bind.
     */
    on<E extends keyof HTMLElementEventMap>(type: E, handler: (target: Wrap<T>, ev: HTMLElementEventMap[E]) => void): Wrap<T> {
        this.el?.addEventListener(type, e => handler(this, e));
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
     * Focus the underlying element.
     */
    focus(): Wrap<T> {
        this.el?.focus();
        return this;
    }

    /**
     * Scroll to the underlying element.
     */
    scrollTo(): Wrap<T> {
        if (this.el) {
            setTimeout(
                () => !this.vertVisible() && this.el.scrollIntoView({block: 'nearest', inline: 'nearest', behavior: 'smooth'}),
                100);
        }
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
     * Run the provided handler in the case there's no underlying element.
     * @param handler Handler to run.
     */
    else(handler: () => void): Wrap<T> {
        if (!this.el) {
            handler();
        }
        return this;
    }

    /**
     * Return a value of the attribute of the underlying element with the given name.
     * @param attrName Attribute name.
     */
    getAttr(attrName: string): string {
        return this.el?.attributes.getNamedItem(attrName)?.value;
    }

    /**
     * Return whether the underlying element is fully visible on the screen along its vertical axis.
     * @private
     */
    private vertVisible(): boolean {
        const r = this.el?.getBoundingClientRect();
        return r && r.top >= 0 && r.bottom <= window.innerWidth;
    }

}
