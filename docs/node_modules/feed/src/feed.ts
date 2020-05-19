import renderAtom from "./atom1";
import renderJSON from "./json";
import renderRSS from "./rss2";
import { FeedOptions, Item, Author, Extension } from "./typings";

export class Feed {
  options: FeedOptions;
  items: Item[] = [];
  categories: string[] = [];
  contributors: Author[] = [];
  extensions: Extension[] = [];

  constructor(options: FeedOptions) {
    this.options = options;
  }

  public addItem = (item: Item) => this.items.push(item);

  public addCategory = (category: string) => this.categories.push(category);

  public addContributor = (contributor: Author) => this.contributors.push(contributor);

  public addExtension = (extension: Extension) => this.extensions.push(extension);

  /**
   * Returns a Atom 1.0 feed
   */
  public atom1 = (): string => renderAtom(this);

  /**
   * Returns a RSS 2.0 feed
   */
  public rss2 = (): string => renderRSS(this);

  /**
   * Returns a JSON1 feed
   */
  public json1 = (): string => renderJSON(this);
}
