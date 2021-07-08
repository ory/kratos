import Tabs from '@theme/Tabs'
import TabItem from '@theme/TabItem'
import React from 'react'
import CodeBlock from '@theme/CodeBlock'
import CodeFromRemote from '@theme/CodeFromRemote'

const FlowContent = ({ item }) => {
  if (item.code) {
    console.warn(item.code, 'asfd')
    return (
      <CodeBlock className={`language-${item.language}`} children={item.code} />
    )
  }

  if (item.image) {
    return <img src={item.image} alt={item.alt} />
  }

  if (item.codeFromRemote) {
    return (
      <CodeFromRemote
        language={item.language}
        link={item.codeFromRemote.link}
        src={item.codeFromRemote.src}
      />
    )
  }

  return <span>Unknown item type: ${JSON.stringify(item)}</span>
}

const GetFlow = ({ items }) => {
  return (
    <Tabs
      defaultValue={Object.keys(items)[0]}
      values={Object.keys(items).map((key) => ({
        label: items[key].label,
        value: key
      }))}
    >
      {Object.keys(items).map((key) => (
        <TabItem key={key} value={key}>
          <FlowContent item={items[key]} />
        </TabItem>
      ))}
    </Tabs>
  )
}

export default GetFlow
