# Colorful

It's not just color, it's everything colorful in terminal.

---------------------

# Color

Color in terminal and only terminal.

![screen shot](./screen-shot.png)

## Programmer

As a programmer, you think they are functions:

```javascript
var color = require('colorful')
color.red('hello')
color.underline('hello')
color.red(color.underline('hello'))
```

## Human

As a human, you think you are a painter:

```javascript
var paint = require('colorful').paint
paint('hello').red.color
paint('hello').bold.underline.red.color
```

**WTF**, is bold, underline a color? If you don't like the idea, try:

```javascript
paint('hello').bold.underline.red.style
```

## Alien

As an alien, you are from outer space, you think it should be:

```javascript
require('colorful').colorful()
'hello'.to.red.color
'hello'.to.underline.bold.red.color
'hello'.to.underline.bold.red.style
```


## Artist

As an artist, you need more colors.

```javascript
var Color = require('colorful').Color;

var s = new Color('colorful');
s.fgcolor = 13;
s.bgcolor = 61;
```

Support ANSI 256 colors. [0 - 255]


## Toxic

Let's posion the string object, just like colors does.

```javascript
require('colorful').toxic()
'hello'.bold
'hello'.red
```


## Detective

As a detective, you think we should detect if color is supported:

```javascript
require('colorful').isSupported

// we can disable color
require('colorful').disabled = true
require('colorful').isSupported
// => false
```

# Colors

- bold
- faint
- italic
- underline
- blink
- overline
- inverse
- conceal
- strike
- black
- black_bg
- red
- red_bg
- green
- green_bg
- yellow
- yellow_bg
- blue
- blue_bg
- magenta
- magenta_bg
- cyan
- cyan_bg
- white
- white_bg
- grey
- gray

# Changelog

**2013-05-22** `2.1.0`

Add toxic API.

**2013-03-22** `2.0.2`

Merge terminal into ansi.

**2013-03-18** `2.0.1`

Add gray color.

**2013-03-18** `2.0.0`

Redesign. Support for ANSI 256 colors.
