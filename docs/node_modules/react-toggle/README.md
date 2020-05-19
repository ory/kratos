An elegant, accessible toggle component for React. Also a glorified checkbox.

<img src="https://camo.githubusercontent.com/7b82df5ece8794631d7b004a6fd1d9fe32a336b6/68747470733a2f2f64337676366c703535716a6171632e636c6f756466726f6e742e6e65742f6974656d732f334132783052335a3245337130523069304531692f53637265656e2532305265636f7264696e67253230323031362d31312d3234253230617425323031312e3433253230414d2e6769663f582d436c6f75644170702d56697369746f722d49643d643661386464343439306336316166646261386130613230383232373361613126763d3631613139333333" height="32px" />

See [usage and examples](http://aaronshaf.github.io/react-toggle/).

## Props

The component takes the following props.

| Prop              | Type       | Description |
|-------------------|------------|-------------|
| `checked`         | _boolean_  | If `true`, the toggle is checked. If `false`, the toggle is unchecked. Use this if you want to treat the toggle as a controlled component |
| `defaultChecked`  | _boolean_  | If `true` on initial render, the toggle is checked. If `false` on initial render, the toggle is unchecked. Use this if you want to treat the toggle as an uncontrolled component |
| `onChange`        | _function_ | Callback function to invoke when the user clicks on the toggle. The function signature should be the following: `function(e) { }`. To get the current checked status from the event, use `e.target.checked`. |
| `onFocus`         | _function_ | Callback function to invoke when field has focus. The function signature should be the following: `function(e) { }` |
| `onBlur`          | _function_ | Callback function to invoke when field loses focus. The function signature should be the following: `function(e) { }` |
| `name`            | _string_   | The value of the `name` attribute of the wrapped \<input\> element |
| `value`           | _string_   | The value of the `value` attribute of the wrapped \<input\> element |
| `id`              | _string_   | The value of the `id` attribute of the wrapped \<input\> element |
| `icons`        | _object_  | If `false`, no icons are displayed. You may also pass custom icon components in `icons={{{checked: <CheckedIcon />, unchecked: <UncheckedIcon />}}` |
| `aria-labelledby` | _string_   | The value of the `aria-labelledby` attribute of the wrapped \<input\> element |
| `aria-label`      | _string_   | The value of the `aria-label` attribute of the wrapped \<input\> element |
| `disabled`        | _boolean_  | If `true`, the toggle is disabled. If `false`, the toggle is enabled |

## Installation

```bash
npm install react-toggle
```

## Usage

If you want the default styling, include the component's [CSS](./style.css) with

```javascript
import "react-toggle/style.css" // for ES6 modules
// or
require("react-toggle/style.css") // for CommonJS
```

## Development

```javascript
npm install
npm run dev
```

## Build

```javascript
npm run build
```

## License

MIT
