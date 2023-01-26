import { Wrap } from './element-wrap';

/**
 * Utility class to facilitate the creation of various UI components.
 */
export class UIToolkit {

    /**
     * Create and return a dialog close button.
     * @param onClick Button's click handler.
     */
    static closeButton(onClick: (e: MouseEvent) => void): Wrap<HTMLButtonElement> {
        return Wrap.new('button')
            .classes('dialog-btn-close')
            .attr({type: 'button', ariaLabel: 'Close'})
            .click(onClick);
    }

    /**
     * Create and return a new popup dialog element.
     * @param onSubmit Form submit handler
     */
    static form(onSubmit: () => void): Wrap<HTMLFormElement> {
        return Wrap.new('form')
            // Form submit event
            .on('submit', e => {
                e.preventDefault();
                onSubmit();
            });
    }

    /**
     * Create and return a new input element.
     */
    static input(name: string, type = 'text', placeholder: string = null, autocomplete: string = null, required?: boolean): Wrap<HTMLInputElement> {
        return Wrap.new('input')
            .classes('input')
            .attr({name, type, placeholder, autocomplete, required: required && 'required', size: '1'});
    }

    /**
     * Create and return a new input container element.
     */
    static inputGroup(): Wrap<HTMLDivElement> {
        return Wrap.new('div').classes('input-group');
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
}
