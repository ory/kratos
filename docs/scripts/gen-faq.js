// read.js
const fs = require('fs');
const yaml = require('js-yaml');

try {
    let fayYaml = fs.readFileSync('./docs/faq.yaml', 'utf8');
    let faq = yaml.safeLoad(fayYaml);

    const taglist = Array.from(new Set(faq.map(el => { return el.tags }).flat(1)))
    

    
    data=`
---
id: faq
title: Faq
---

export const Question = ({children, tags}) => (
    <div class={{tags}}>
      {children}
    </div>
  );

${taglist.join(" ") }

import Faq from '@theme/Faq'

<Faq />

`

    
    

    faq.forEach(el => {
        data += `<Question tags="${el.tags.map( tag => {return tag }).join(" ")}">\n`
        data += `M: ${el.tags.map( tag => {return "#"+tag }).join(" ")} \n` 
        data += `Q: ${el.q}\n` 
        data += `A: ${el.a}\n`
        data += `</Question>\n\n`
    });

    const faqMdx = require('fs') 
    // Write data in 'Output.txt' . 
    fs.writeFile('./docs/docs/faq.mdx', data, (err) => { 
        
        // In case of a error throw err. 
        if (err) throw err; 
    }) 

    console.log(taglist)

} catch (e) {
    console.log(e);
}