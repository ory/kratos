import Tabs from '@theme/Tabs'
import TabItem from '@theme/TabItem'
import React from "react";
import CodeBlock from '@theme/CodeBlock'

const GetFlow = ({items}) => (
  <Tabs
    defaultValue={Object.keys(items)[0]}
    values={Object.keys(items).map((key) => ({
      label: items[key].label,
      value: key
    }))}>
    {Object.keys(items).map((key) => (
      <TabItem key={key} value={key}>
        {items[key].code ?
          <CodeBlock className={`language-${items[key].language}`} children={items[key].code}/>
          : <img src={items[key].image} alt={items[key].alt}/>}
      </TabItem>
    ))}
  </Tabs>
)

export default GetFlow
