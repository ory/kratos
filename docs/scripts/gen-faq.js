// read.js
const fs = require('fs');
const yaml = require('js-yaml');
const { Remarkable } = require('remarkable');


try {
    // Generating FAQ.mdx

    let fayYaml = fs.readFileSync('./docs/faq.yaml', 'utf8');
    let faq = yaml.safeLoad(fayYaml);


    const tags = Array.from(new Set(faq.map(el => { return el.tags }).flat(1)))

    
    data=`---
id: faq
title: Faq
---



import {Question, Faq} from '@theme/Faq'

<Faq tags="${tags.join(' ')}"/>
<br/><br/>

`
    md = new Remarkable();
    faq.forEach(el => {
        react_tags = el.tags.map((tag) => {return tag+"_src-theme-"})
        data += `<Question tags="question_src-theme- ${react_tags.join(" ")}">\n`
        data += `${el.tags.map( tag => {return "#"+tag }).join(" ")} <br/>\n` 
        data += md.render(`**Q**: ${el.q}`) 
        data += md.render(`**A**: ${el.a}\n`)
        if (el.context) {
            data += md.render(`context: ${el.context}\n`)
        }
        data += `</Question>\n\n<br/>`
    });

    fs.writeFile('./docs/docs/faq.mdx', data, (err) => { 
        if (err) throw err; 
    }) 

    // Generating faq.module.css
    const taglist = Array.from(new Set(faq.map(el => { return el.tags }).flat(1)))
    css_file=`
.pills,
.tabs {
    font-weight:var(--ifm-font-weight-bold)
}
.pills {
    padding-left:0
}
.pills__item {
    border-radius:.5rem;
    cursor:pointer;
    display:inline-block;
    padding:.25rem 1rem;
    transition:background var(--ifm-transition-fast) var(--ifm-transition-timing-default)
}
.pills__item--active {
    background:var(--ifm-pills-color-background-active);
    color:var(--ifm-pills-color-active)
}
.pills__item:not(.pills__item--active):hover {
    background-color:var(--ifm-pills-color-background-active)
}
.pills__item:not(:first-child) {
    margin-left:var(--ifm-pills-spacing)
}
.pills__item:not(:last-child) {
    margin-right:var(--ifm-pills-spacing)
}
.pills--block {
    display:flex;
    justify-content:stretch
}
.pills--block .pills__item {
    flex-grow:1;
    text-align:center
}
.tabs {
    display:flex;
    overflow-x:auto;
    color:var(--ifm-tabs-color);
    margin-bottom:0;
    padding-left:0
}
.tabs__item {
    border-bottom:3px solid transparent;
    border-radius:var(--ifm-global-radius);
    cursor:pointer;
    display:inline-flex;
    padding:var(--ifm-tabs-padding-vertical) var(--ifm-tabs-padding-horizontal);
    margin:0;
    transition:background-color var(--ifm-transition-fast) var(--ifm-transition-timing-default)
}
.tabs__item--active {
    border-bottom-color:var(--ifm-tabs-color-active);
    border-bottom-left-radius:0;
    border-bottom-right-radius:0;
    color:var(--ifm-tabs-color-active)
}
.tabs__item:hover {
    background-color:var(--ifm-hover-overlay)
}
.tabs--block {
    justify-content:stretch
}
.tabs--block .tabs__item {
    flex-grow:1;
    justify-content:center
}


p {
    margin-bottom: 0px;
}
    
.selected {
    background-color: #ffba00;
}

div.question {
    display: none;
}
`
    taglist.forEach( tag => {
        css_file += `
li.selected.${tag} {
    color:red;
}

li.selected.${tag}~.question.${tag} {
    display: inline;
    
}
`
    })
    fs.writeFile('./docs/src/theme/faq.module.css', css_file, (err) => { 
        if (err) throw err; 
    }) 

} catch (e) {
    throw e
}
    