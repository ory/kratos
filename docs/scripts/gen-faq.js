// read.js
const fs = require('fs')
const yaml = require('js-yaml')
const { Remarkable } = require('remarkable')
const path = require('path')
const yamlPath = path.resolve('./docs/faq.yaml')

// Generating FAQ.mdx

if (!fs.existsSync(yamlPath)) {
  //file exists
  console.warn('.yaml File does not exists, skipping generating FAQ')
  return 0
}

let faqYaml = fs.readFileSync(yamlPath, 'utf8')
let faq = yaml.load(faqYaml)

const tags = Array.from(new Set(faq.map(({ tags }) => tags).flat(1)))

let data = `---
id: faq
title: Frequently Asked Questions (FAQ)
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
    let css_file=``

taglist.forEach((tag) => {
  css_file += `
li.selected.${tag} {
    color:red;
}

li.selected.${tag}~.question.${tag} {
    display: inline;
    
}
`
})

fs.writeFile('./docs/src/theme/faq.module.gen.css', css_file, (err) => { 
    if (err) throw err; 
}) 