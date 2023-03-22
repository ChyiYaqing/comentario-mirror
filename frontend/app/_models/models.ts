export interface Language {
    /** Name of the language in that language. */
    nativeName: string;
    /** Two-letter ISO 639-1 language code. */
    code: string;
    /** Language weight to order languages by. */
    weight: number;
    /** Date format for the language. */
    dateFormat: string;
    /** Datetime format for the language. */
    datetimeFormat: string;
}
