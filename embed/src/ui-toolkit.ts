import { Wrap } from './element-wrap';

/**
 * Utility class to facilitate the creation of various UI components.
 */
export class UIToolkit {

    /**
     * Create and return a new div element.
     * @param classes Classes to add to the div.
     */
    static div(...classes: string[]): Wrap<HTMLDivElement> {
        return Wrap.new('div').classes(...classes);
    }

    /**
     * Create and return a dialog close button.
     * @param onClick Button's click handler.
     */
    static closeButton(onClick: () => void): Wrap<HTMLButtonElement> {
        return Wrap.new('button')
            .classes('dialog-btn-close')
            .attr({type: 'button', ariaLabel: 'Close'})
            .click(onClick);
    }

    /**
     * Create and return a new popup dialog element.
     * @param onSubmit Form submit handler.
     * @param onCancel Form cancel handler (triggered when the user presses Esc).
     */
    static form(onSubmit: () => void, onCancel?: () => void): Wrap<HTMLFormElement> {
        const submit = (f: Wrap<HTMLFormElement>, e: Event) => {
            // Prevent default handling
            e.preventDefault();

            // Mark all inputs touched to show their validation
            [...f.element.getElementsByTagName('input'), ...f.element.getElementsByTagName('textarea')]
                .forEach(el => new Wrap(el).classes('touched'));

            // Run the submit handler if the form is valid
            if (f.element.checkValidity()) {
                onSubmit();
            }
        };
        const cancel = (e: Event) => {
            // If there's a cancel handler provided
            if (onCancel) {
                e.preventDefault();
                onCancel();
            }
        };
        return Wrap.new('form')
            // Intercept form submit event
            .on('submit', submit)
            // Submit the form on Ctrl+Enter and cancel with Esc
            .on('keydown', (f, e) => {
                switch (e.code) {
                    case 'Enter':
                        return e.ctrlKey && !e.shiftKey && !e.altKey && !e.metaKey && submit(f, e);
                    case 'Escape':
                        return !e.ctrlKey && !e.shiftKey && !e.altKey && !e.metaKey && cancel(e);
                }
            });
    }

    /**
     * Create and return a new input element.
     */
    static input(name: string, type = 'text', placeholder: string | null | undefined = null, autocomplete: string | null | undefined = null, required?: boolean): Wrap<HTMLInputElement> {
        return Wrap.new('input')
            .classes('input')
            .attr({name, type, placeholder, autocomplete, required: required ? 'required' : null, size: '1'})
            // Add the touched class on blur, which is used to highlight invalid input
            .on('blur', t => t.classes('touched'));
    }

    /**
     * Create and return a new textarea element.
     */
    static textarea(placeholder:  null | undefined | string, required: boolean, autoExpand: boolean): Wrap<HTMLTextAreaElement> {
        return Wrap.new('textarea')
            .attr({placeholder, required: required ? 'required' : null})
            // Add the touched class on blur, which is used to highlight invalid input
            .on('blur', t => t.classes('touched'))
            // Enable automatic height adjusting on input, if needed
            .on('input', t =>
                autoExpand &&
                t.style('height:auto')
                    .style(`height:${Math.min(Math.max(t.element.scrollHeight + t.element.offsetHeight - t.element.clientHeight, 75), 400)}px`));
    }

    /**
     * Create and return a new button element.
     * @param label Label of the button (HTML).
     * @param onClick Button's click handler.
     * @param classes Additional button classes to add.
     */
    static button(label: string, onClick?: (btn: Wrap<HTMLButtonElement>, e: MouseEvent) => void,  ...classes: (false | null | undefined | string)[]): Wrap<HTMLButtonElement> {
        return Wrap.new('button').classes('button', ...classes).html(label).attr({type: 'button'}).click(onClick);
    }

    /**
     * Create and return a new submit button element.
     * @param title Title of the button, and, if glyph is false, also its label.
     * @param glyph Whether to draw a "carriage return" glyph instead of text.
     */
    static submit(title: string, glyph: boolean): Wrap<HTMLButtonElement> {
        return Wrap.new('button')
            .classes('button', 'submit-button', glyph && 'submit-glyph')
            .inner(glyph ? '' : title)
            .attr({type: 'submit', title});
    }

    /**
     * Return a textual representation of an SVG image.
     * @param width Viewport width.
     * @param height Viewport height.
     * @param content SVG content.
     */
    static svg(width: number, height: number, content: string): string {
        return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${width} ${height}">${content}</svg>`;
    }
}
