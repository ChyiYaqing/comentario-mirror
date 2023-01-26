import { createPopper } from '@popperjs/core/lib/popper-lite';
import { arrow, flip, offset, preventOverflow } from '@popperjs/core';
import { Placement } from '@popperjs/core/lib/enums';
import { Wrap } from './element-wrap';
import { UIToolkit } from './ui-toolkit';

export interface DialogPositioning {
    /** Reference element. */
    ref: Wrap<any>;
    /** Dialog placement. */
    placement?: Placement;
}

/**
 * Generic popup dialog component.
 */
export class Dialog {

    private backdrop: Wrap<HTMLDivElement>;
    private dialogBox: Wrap<HTMLDivElement>;
    private resolve: () => void;
    private animationDone?: () => void;

    _confirmed = false;

    constructor(
        /** Parent element that will host the dialog and the backdrop. */
        private readonly parent: Wrap<any>,
        /** Dialog title. */
        private readonly title: string,
        /** Optional positioning settings for the dialog. */
        private readonly pos?: DialogPositioning,
    ) {}

    /**
     * Whether the dialog has been confirmed when closed.
     */
    get confirmed(): boolean {
        return this._confirmed;
    }

    /**
     * Main method that show the dialog and resolves with the provided data when the dialog is closed.
     * @param data Data to resolve the promise with.
     */
    run<T>(data?: T): Promise<T> {
        return new Promise(resolve => {
            this.resolve = () => resolve(data);

            // Create a login box
            this.dialogBox = Wrap.new('div')
                .classes('dialog', 'fade-in')
                .attr({role: 'dialog'})
                // Don't propagate the click to prevent cancelling the dialog, which happens when the click reaches the
                // parent container
                .click(e => e.stopPropagation())
                // Invoke the animation callback when it's either ended or interrupted
                .on('animationend',    () => this.animationDone?.())
                .on('animationcancel', () => this.animationDone?.())
                .append(
                    // Dialog header
                    this.renderHeader(),
                    // Dialog body + contents
                    Wrap.new('div').classes('dialog-body').append(this.renderContent()));

            // Create a backdrop
            this.backdrop = Wrap.new('div')
                .classes('backdrop')
                // Cancel the dialog when clicked outside
                .click(() => this.dismiss())
                .appendTo(this.parent);

            // Set up the animation end callback
            this.animationDone = () => {
                this.animationDone = null;

                // Scroll to the dialog element, if necessary
                this.dialogBox.scrollTo();

                // Call the callback
                this.onShow();
            };

            // Insert the dialog into the DOM
            this.dialogBox.appendTo(this.parent);

            // Position the element using Popper, if required
            this.popperBind();
        });
    }

    /**
     * Must render and return the content of the dialog.
     * @protected
     */
    protected renderContent(): Wrap<any> {
        return null;
    }

    /**
     * Dismiss the dialog, setting the confirmed property.
     * @param confirmed Value of confirmed to set.
     * @protected
     */
    protected dismiss(confirmed?: boolean) {
        // Set up the animation end callback
        this.animationDone = () => {
            this.animationDone = null;

            // Clean up the elements
            this.dialogBox.remove();
            this.backdrop.remove();

            // Resolve the promise, returning the dialog
            this._confirmed = !!confirmed;
            this.resolve();
        };

        // Animate-close the dialog
        this.dialogBox.noClasses('fade-in').classes('fade-out');
        this.backdrop.classes('fade-out');
    }

    /**
     * Called whenever the dialog has been shown.
     * @protected
     */
    protected onShow(): void {
        // Does nothing by default
    }

    private renderHeader(): Wrap<HTMLDivElement> {
        return Wrap.new('div')
            .classes('dialog-header')
            // Title
            .inner(this.title)
            // Close button
            .append(UIToolkit.closeButton(() => this.dismiss()));
    }

    private popperBind() {
        // Position the element using Popper, if required
        if (!this.pos?.ref?.ok) {
            return;
        }

        // Add an arrow element to the dialog
        const wa = this.pos?.ref?.ok && Wrap.new('div').classes('dialog-arrow');
        this.dialogBox.append(wa);

        // Set up the arrow modifier
        const modArrow = arrow;
        modArrow.options = {element: wa.element, padding: 7};

        // Set up the offset modifier
        const modOffset = offset;
        const horz = !!this.pos?.placement?.match(/left|right/); // True when the dialog is on left/right side of the ref
        modOffset.options = {offset: [horz ? 8 : 0, horz ? 0 : 8]};

        createPopper(
            this.pos.ref.element,
            this.dialogBox.element,
            {
                placement: this.pos.placement,
                modifiers: [preventOverflow, flip, modArrow, modOffset],
            });
    }
}
