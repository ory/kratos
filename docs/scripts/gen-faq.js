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
button.selected.${tag} {
    color:red;
}

button.selected.${tag}~.question.${tag} {
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
    