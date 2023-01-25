import { Wrap } from './element-wrap';
import { Dialog } from './dialog';

export class MarkdownHelp extends Dialog {

    constructor(parent: Wrap<any>) {
        super(parent, 'Markdown help');
    }

    /**
     * Instantiate and show the dialog. Return a promise that resolves as soon as the dialog is closed.
     * @param parent Parent element for the dialog.
     */
    static run(parent: Wrap<any>): void {
        new MarkdownHelp(parent).run();
    }

    override renderContent(): Wrap<any> {
        return Wrap.new('table')
            .classes('markdown-help')
            .append(
                Wrap.new('tr')
                    .append(
                        Wrap.new('td').html('<i>italics</i>'),
                        Wrap.new('td').html('surround text with <pre>*asterisks*</pre>')),
                Wrap.new('tr')
                    .append(
                        Wrap.new('td').html('<b>bold</b>'),
                        Wrap.new('td').html('surround text with <pre>**two asterisks**</pre>')),
                Wrap.new('tr')
                    .append(
                        Wrap.new('td').html('<pre>code</pre>'),
                        Wrap.new('td').html('surround text with <pre>`backticks`</pre>')),
                Wrap.new('tr')
                    .append(
                        Wrap.new('td').html('<del>strikethrough</del>'),
                        Wrap.new('td').html('surround text with <pre>~~two tilde characters~~</pre>')),
                Wrap.new('tr')
                    .append(
                        Wrap.new('td').html('<a href="https://example.com">hyperlink</a>'),
                        Wrap.new('td').html('<pre>[hyperlink](https://example.com)</pre> or just a bare URL')),
                Wrap.new('tr')
                    .append(
                        Wrap.new('td').html('<blockquote>quote</blockquote>'),
                        Wrap.new('td').html('prefix with <pre>&gt;</pre>')));
    }
}
