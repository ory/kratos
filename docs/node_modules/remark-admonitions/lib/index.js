// handles different types of whitespace
const unified = require("unified");
const html = require("rehype-parse");
const visit = require("unist-util-visit");

const NEWLINE = "\n";

// natively supported types
const types = {
  // aliases for infima color names
  secondary: "note",
  info: "important",
  success: "tip",
  danger: "warning",
  // base types
  note: {
    ifmClass: "secondary",
    keyword: "note",
    emoji: "ℹ️", // '&#x2139;'
    svg:
      '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="16" viewBox="0 0 14 16"><path fill-rule="evenodd" d="M6.3 5.69a.942.942 0 0 1-.28-.7c0-.28.09-.52.28-.7.19-.18.42-.28.7-.28.28 0 .52.09.7.28.18.19.28.42.28.7 0 .28-.09.52-.28.7a1 1 0 0 1-.7.3c-.28 0-.52-.11-.7-.3zM8 7.99c-.02-.25-.11-.48-.31-.69-.2-.19-.42-.3-.69-.31H6c-.27.02-.48.13-.69.31-.2.2-.3.44-.31.69h1v3c.02.27.11.5.31.69.2.2.42.31.69.31h1c.27 0 .48-.11.69-.31.2-.19.3-.42.31-.69H8V7.98v.01zM7 2.3c-3.14 0-5.7 2.54-5.7 5.68 0 3.14 2.56 5.7 5.7 5.7s5.7-2.55 5.7-5.7c0-3.15-2.56-5.69-5.7-5.69v.01zM7 .98c3.86 0 7 3.14 7 7s-3.14 7-7 7-7-3.12-7-7 3.14-7 7-7z"/></svg>'
  },
  tip: {
    ifmClass: "success",
    keyword: "tip",
    emoji: "💡", //'&#x1F4A1;'
    svg:
      '<svg xmlns="http://www.w3.org/2000/svg" width="12" height="16" viewBox="0 0 12 16"><path fill-rule="evenodd" d="M6.5 0C3.48 0 1 2.19 1 5c0 .92.55 2.25 1 3 1.34 2.25 1.78 2.78 2 4v1h5v-1c.22-1.22.66-1.75 2-4 .45-.75 1-2.08 1-3 0-2.81-2.48-5-5.5-5zm3.64 7.48c-.25.44-.47.8-.67 1.11-.86 1.41-1.25 2.06-1.45 3.23-.02.05-.02.11-.02.17H5c0-.06 0-.13-.02-.17-.2-1.17-.59-1.83-1.45-3.23-.2-.31-.42-.67-.67-1.11C2.44 6.78 2 5.65 2 5c0-2.2 2.02-4 4.5-4 1.22 0 2.36.42 3.22 1.19C10.55 2.94 11 3.94 11 5c0 .66-.44 1.78-.86 2.48zM4 14h5c-.23 1.14-1.3 2-2.5 2s-2.27-.86-2.5-2z"/></svg>'
  },
  warning: {
    ifmClass: "danger",
    keyword: "warning",
    emoji: "🔥", //'&#x1F525;'
    svg:
      '<svg xmlns="http://www.w3.org/2000/svg" width="12" height="16" viewBox="0 0 12 16"><path fill-rule="evenodd" d="M5.05.31c.81 2.17.41 3.38-.52 4.31C3.55 5.67 1.98 6.45.9 7.98c-1.45 2.05-1.7 6.53 3.53 7.7-2.2-1.16-2.67-4.52-.3-6.61-.61 2.03.53 3.33 1.94 2.86 1.39-.47 2.3.53 2.27 1.67-.02.78-.31 1.44-1.13 1.81 3.42-.59 4.78-3.42 4.78-5.56 0-2.84-2.53-3.22-1.25-5.61-1.52.13-2.03 1.13-1.89 2.75.09 1.08-1.02 1.8-1.86 1.33-.67-.41-.66-1.19-.06-1.78C8.18 5.31 8.68 2.45 5.05.32L5.03.3l.02.01z"/></svg>'
  },
  important: {
    ifmClass: "info",
    keyword: "important",
    emoji: "❗️", //'&#x2757;'
    svg:
      '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="16" viewBox="0 0 14 16"><path fill-rule="evenodd" d="M7 2.3c3.14 0 5.7 2.56 5.7 5.7s-2.56 5.7-5.7 5.7A5.71 5.71 0 0 1 1.3 8c0-3.14 2.56-5.7 5.7-5.7zM7 1C3.14 1 0 4.14 0 8s3.14 7 7 7 7-3.14 7-7-3.14-7-7-7zm1 3H6v5h2V4zm0 6H6v2h2v-2z"/></svg>'
  },
  caution: {
    ifmClass: "warning",
    keyword: "caution",
    emoji: "⚠️", // '&#x26A0;&#xFE0F;'
    svg:
      '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16"><path fill-rule="evenodd" d="M8.893 1.5c-.183-.31-.52-.5-.887-.5s-.703.19-.886.5L.138 13.499a.98.98 0 0 0 0 1.001c.193.31.53.501.886.501h13.964c.367 0 .704-.19.877-.5a1.03 1.03 0 0 0 .01-1.002L8.893 1.5zm.133 11.497H6.987v-2.003h2.039v2.003zm0-3.004H6.987V5.987h2.039v4.006z"/></svg>'
  }
};

// default options for plugin
const defaultOptions = {
  customTypes: [],
  useDefaultTypes: true,
  infima: true,
  tag: ":::",
  icons: "svg"
};

// override default options
const configure = options => {
  const { customTypes, ...baseOptions } = {
    ...defaultOptions,
    ...options
  };

  return {
    ...baseOptions,
    types: { ...types, ...customTypes }
  };
};

// escape regex special characters
function escapeRegExp(s) {
  return s.replace(new RegExp(`[-[\\]{}()*+?.\\\\^$|/]`, "g"), "\\$&");
}

// helper: recursively replace nodes
const _nodes = ({
  tagName: hName,
  properties: hProperties,
  position,
  children
}) => {
  return {
    type: "admonitionHTML",
    data: {
      hName,
      hProperties
    },
    position,
    children: children.map(_nodes)
  };
};

// convert HTML to MDAST (must be a single root element)
const nodes = markup => {
  return _nodes(
    unified()
      .use(html)
      .parse(markup).children[0].children[1].children[0]
  );
};

// create a simple text node
const text = value => {
  return {
    type: "text",
    value
  };
};

// create a node that will compile to HTML
const element = (tagName, classes = [], children = []) => {
  return {
    type: "admonitionHTML",
    data: {
      hName: tagName,
      hProperties: classes.length
        ? {
            className: classes
          }
        : {}
    },
    children
  };
};

// passed to unified.use()
// you have to use a named function for access to `this` :(
module.exports = function attacher(options) {
  const config = configure(options);

  // match to determine if the line is an opening tag
  const keywords = Object.keys(config.types)
    .map(escapeRegExp)
    .join("|");
  const tag = escapeRegExp(config.tag);
  const regex = new RegExp(`${tag}(${keywords})(?: *(.*))?\n`);
  const escapeTag = new RegExp(escapeRegExp(`\\${config.tag}`), "g");

  // the tokenizer is called on blocks to determine if there is an admonition present and create tags for it
  function blockTokenizer(eat, value, silent) {
    // stop if no match or match does not start at beginning of line
    const match = regex.exec(value);
    if (!match || match.index !== 0) return false;
    // if silent return the match
    if (silent) return true;

    const now = eat.now();
    const [opening, keyword, title] = match;
    const food = [];
    const content = [];

    // consume lines until a closing tag
    let idx = 0;
    while ((idx = value.indexOf(NEWLINE)) !== -1) {
      // grab this line and eat it
      const next = value.indexOf(NEWLINE, idx + 1);
      const line =
        next !== -1 ? value.slice(idx + 1, next) : value.slice(idx + 1);
      food.push(line);
      value = value.slice(idx + 1);
      // the closing tag is NOT part of the content
      if (line.startsWith(config.tag)) break;
      content.push(line);
    }

    // consume the processed tag and replace escape sequences
    const contentString = content.join(NEWLINE).replace(escapeTag, config.tag);
    const add = eat(opening + food.join(NEWLINE));

    // parse the content in block mode
    const exit = this.enterBlock();
    const contentNodes = element(
      "div",
      "admonition-content",
      this.tokenizeBlock(contentString, now)
    );
    exit();
    // parse the title in inline mode
    const titleNodes = this.tokenizeInline(title || keyword, now);
    // create the nodes for the icon
    const entry = config.types[keyword];
    const settings = typeof entry === "string" ? config.types[entry] : entry;
    const iconNodes =
      config.icons === "svg" ? nodes(settings.svg) : text(settings.emoji);
    const iconContainerNodes =
      config.icons === "none"
        ? []
        : [element("span", "admonition-icon", [iconNodes])];

    // build the nodes for the full markup
    const admonition = element(
      "div",
      ["admonition", `admonition-${keyword}`].concat(
        settings.ifmClass ? ["alert", `alert--${settings.ifmClass}`] : []
      ),
      [
        element("div", "admonition-heading", [
          element("h5", "", iconContainerNodes.concat(titleNodes))
        ]),
        contentNodes
      ]
    );

    return add(admonition);
  }

  // add tokenizer to parser after fenced code blocks
  const Parser = this.Parser.prototype;
  Parser.blockTokenizers.admonition = blockTokenizer;
  Parser.blockMethods.splice(
    Parser.blockMethods.indexOf("fencedCode") + 1,
    0,
    "admonition"
  );
  Parser.interruptParagraph.splice(
    Parser.interruptParagraph.indexOf("fencedCode") + 1,
    0,
    ["admonition"]
  );
  Parser.interruptList.splice(
    Parser.interruptList.indexOf("fencedCode") + 1,
    0,
    ["admonition"]
  );
  Parser.interruptBlockquote.splice(
    Parser.interruptBlockquote.indexOf("fencedCode") + 1,
    0,
    ["admonition"]
  );

  // TODO: add compiler rules for converting back to markdown

  return function transformer(tree) {
    // escape everything except admonitionHTML nodes
    visit(
      tree,
      node => {
        return node.type !== "admonitionHTML";
      },
      function visitor(node) {
        if (node.value) node.value = node.value.replace(escapeTag, config.tag);
        return node;
      }
    );
  };
};
