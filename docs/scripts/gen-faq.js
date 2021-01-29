// read.js
const fs = require('fs');
const yaml = require('js-yaml');

try {
    let fayYaml = fs.readFileSync('./docs/faq.yaml', 'utf8');
    let faq = yaml.safeLoad(fayYaml);

    const taglist = Array.from(new Set(faq.map(el => { return el.tags }).flat(1)))
    

    
    data=`---
id: faq
title: Faq
---

export const Question = ({children, tags}) => (
    <div class={tags} style="display: block;">
      {children}
    </div>
  );


${taglist.join(" ") }
bla

import Faq from '@theme/Faq'

<Faq />

`

    
    

    faq.forEach(el => {
        data += `<Question tags="${el.tags.join(" ")}">\n`
        data += `M: ${el.tags.map( tag => {return "#"+tag }).join(" ")} \n` 
        data += `Q: ${el.q}\n` 
        data += `A: ${el.a}\n`
        data += `</Question>\n\n`
    });

    fs.writeFile('./docs/docs/faq.mdx', data, (err) => { 
        if (err) throw err; 
    }) 
} catch (e) {
    console.log(e);
}