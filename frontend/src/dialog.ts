import { Wrap } from './element-wrap';
import { UIToolkit } from './ui-toolkit';

/**
 * Generic popup dialog component.
 */
export class Dialog {

    private backdrop: Wrap<HTMLDivElement>;
    private container: Wrap<HTMLDivElement>;
    private dialogBox: Wrap<HTMLDivElement>;
    private resolve: () => void;
    private animationDone?: () => void;

    _confirmed = false;

    constructor(
        private readonly parent: Wrap<any>,
        private readonly title: string,
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
            this.backdrop = Wrap.new('div').classes('backdrop').appendTo(this.parent);

            // Create a full-size background container on top of the backdrop
            this.container = Wrap.new('div')
                .classes('dialog-container')
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

            // Add the dialog to the container
            this.dialogBox.appendTo(this.container);
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
            this.container.remove();
            this.backdrop.remove();

            // Resolve the promise, returning the dialog
            this._confirmed = !!confirmed;
            this.resolve();
        };

        // Animate-close the dialog
        this.dialogBox.noClasses('fade-in').classes('fade-out');
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
}
