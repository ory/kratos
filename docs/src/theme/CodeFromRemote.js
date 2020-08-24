import React, {useEffect, useState} from 'react'
import fetch from 'node-fetch'
import CodeBlock from '@theme/CodeBlock'

const CodeFromRemote = ({src,link}) => {
  const [content, setContent] = useState('')

  useEffect(() => {
    fetch(src).then(body => body.text()).then(setContent).catch(console.err)
  }, [])

  return (
    <>
      <CodeBlock metastring={link && `title="${link}"`} className={"language-typescript"} children={content}/>
    </>
  )
}

export default CodeFromRemote
