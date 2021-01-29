// read.js
const fs = require('fs');
const yaml = require('js-yaml');



try {
    // Generating FAQ.mdx

    let fayYaml = fs.readFileSync('./docs/faq.yaml', 'utf8');
    let faq = yaml.safeLoad(fayYaml);


    

    
    data=`---
id: faq
title: Faq
---

export const Question = ({children, tags}) => (
    <div className={tags}>
      {children}
    </div>
  );

import Faq from '@theme/Faq'

<Faq />

`

    faq.forEach(el => {
        react_tags = el.tags.map((tag) => {return tag+"_src-theme-"})
        data += `<Question tags="question_src-theme- ${react_tags.join(" ")}">\n`
        data += `M: ${el.tags.map( tag => {return "#"+tag }).join(" ")} \n` 
        data += `Q: ${el.q}\n` 
        data += `A: ${el.a}\n`
        data += `</Question>\n\n`
    });

    fs.writeFile('./docs/docs/faq.mdx', data, (err) => { 
        if (err) throw err; 
    }) 

    // Generating faq.module.css
    const taglist = Array.from(new Set(faq.map(el => { return el.tags }).flat(1)))
    css_file=`
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
    