import { Wrap } from './element-wrap';
import { Dialog, DialogPositioning } from './dialog';

export class MarkdownHelp extends Dialog {

    constructor(parent: Wrap<any>, pos: DialogPositioning) {
        super(parent, 'Markdown help', pos);
    }

    /**
     * Instantiate and show the dialog. Return a promise that resolves as soon as the dialog is closed.
     * @param parent Parent element for the dialog.
     * @param pos Positioning options.
     */
    static run(parent: Wrap<any>, pos: DialogPositioning): void {
        new MarkdownHelp(parent, pos).run();
    }

    override renderContent(): Wrap<any> {
        return Wrap.new('div')
            .classes('table-container')
            .append(
                Wrap.new('table')
                    .classes('table')
                    .append(
                        this.row('<i>italics</i>',                              'surround text with <pre>*asterisks*</pre>'),
                        this.row('<b>bold</b>',                                 'surround text with <pre>**two asterisks**</pre>'),
                        this.row('<pre>code</pre>',                             'surround text with <pre>`backticks`</pre>'),
                        this.row('<del>strikethrough</del>',                    'surround text with <pre>~~two tilde characters~~</pre>'),
                        this.row('<a href="https://example.com">hyperlink</a>', '<pre>[hyperlink](https://example.com)</pre> or just a bare URL'),
                        this.row('<blockquote>quote</blockquote>',              'prefix with <pre>&gt;</pre>')));
    }

    private row(md: string, text: string): Wrap<any> {
        return Wrap.new('tr')
            .append(Wrap.new('td').html(md), Wrap.new('td').html(text));
    }
}
