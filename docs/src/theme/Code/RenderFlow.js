import Tabs from '@theme/Tabs'
import TabItem from '@theme/TabItem'
import React from 'react'
import CodeFromRemote from '../CodeFromRemote'

import loginBrowser from './images/login/browser.png'
import settingsBrowser from './images/settings/browser.png'
import registrationBrowser from './images/registration/browser.png'
import recoveryBrowser from './images/recovery/browser.png'

const ucfirst = (name) => name.charAt(0).toUpperCase() + name.slice(1)

const RenderFlow = ({ flow }) => {
  let image
  switch (flow) {
    case 'registration':
      image = (
        <img src={registrationBrowser} alt="User Registration HTML Form" />
      )
      break
    case 'login':
      image = <img src={loginBrowser} alt="User Login HTML Form" />
      break
    case 'settings':
      image = <img src={settingsBrowser} alt="Profile Settings HTML Form" />
      break
    case 'recovery':
      image = <img src={recoveryBrowser} alt="Account Recovery HTML Form" />
      break
    case 'verification':
      image = <img src={recoveryBrowser} alt="Email Verification HTML Form" />
      break
  }

  return (
    <Tabs
      defaultValue="browser"
      values={[
        { label: 'Browser UI', value: 'browser' },
        { label: 'Golang (API Flow)', value: 'golang' },
        { label: 'ExpressJS', value: 'express' },
        { label: 'ReactJS', value: 'react' },
        { label: 'React Native', value: 'react-native' }
      ]}
    >
      <TabItem value="browser">{image}</TabItem>
      <TabItem value="express">
        <CodeFromRemote
          lang="js"
          link={`https://github.com/ory/kratos-selfservice-ui-node/blob/master/src/routes/${flow}.ts`}
          src={`https://raw.githubusercontent.com/ory/kratos-selfservice-ui-node/master/src/routes/${flow}.ts`}
        />
        The views can be rather simple, as Ory Kratos provides you with all the
        information you need for rendering the forms. The following examples use
        Handlebars and a generic form generator to render the Flow:
        <Tabs
          defaultValue="view"
          values={[
            { label: `${ucfirst(flow)} View`, value: 'view' },
            { label: 'Generic Form View', value: 'generic-form' },
            { label: 'Example Input Form Element', value: 'input-form' }
          ]}
        >
          <TabItem value="view">
            <CodeFromRemote
              lang="handlebars"
              link={`https://github.com/ory/kratos-selfservice-ui-node/blob/master/views/${flow}.hbs`}
              src={`https://raw.githubusercontent.com/ory/kratos-selfservice-ui-node/master/views/${flow}.hbs`}
            />
          </TabItem>
          <TabItem value="generic-form">
            <CodeFromRemote
              lang="handlebars"
              link="https://github.com/ory/kratos-selfservice-ui-node/blob/master/views/partials/ui.hbs"
              src="https://raw.githubusercontent.com/ory/kratos-selfservice-ui-node/master/views/partials/ui.hbs"
            />
          </TabItem>
          <TabItem value="input-form">
            <CodeFromRemote
              lang="handlebars"
              link="https://github.com/ory/kratos-selfservice-ui-node/blob/master/views/partials/ui_node_input_default.hbs"
              src="https://raw.githubusercontent.com/ory/kratos-selfservice-ui-node/master/views/partials/ui_node_input_default.hbs"
            />
          </TabItem>
        </Tabs>
        The rest of the form partials can be found{' '}
        <a href="https://github.com/ory/kratos-selfservice-ui-node/tree/master/views/partials">
          here
        </a>
        .
      </TabItem>
      <TabItem value="react">
        A React example is currently in the making.
      </TabItem>
      <TabItem value="golang">
        <CodeFromRemote
          lang="tsx"
          link={`https://github.com/ory/kratos/blob/master/examples/go/selfservice/${flow}/main.go.tsx`}
          src={`https://raw.githubusercontent.com/ory/kratos/master/examples/go/selfservice/${flow}/main.go`}
        />
        The rest of the code examples partials can be found{' '}
        <a href="https://github.com/ory/kratos/blob/master/examples">here</a>
      </TabItem>
      <TabItem value="react-native">
        <CodeFromRemote
          lang="tsx"
          link={`https://github.com/ory/kratos-selfservice-ui-react-native/blob/master/src/components/Routes/${ucfirst(
            flow
          )}.tsx`}
          src={`https://raw.githubusercontent.com/ory/kratos-selfservice-ui-react-native/master/src/components/Routes/${ucfirst(
            flow
          )}.tsx`}
        />
        The views can be rather simple, as Ory Kratos provides you with all the
        information you need for rendering the forms.
        <Tabs
          defaultValue="generic-form"
          values={[
            { label: 'Generic Form View', value: 'generic-form' },
            { label: 'Example Input Form Element', value: 'input-form' }
          ]}
        >
          <TabItem value="generic-form">
            <CodeFromRemote
              lang="handlebars"
              link="https://github.com/ory/kratos-selfservice-ui-react-native/blob/master/src/components/Form/Form.tsx"
              src="https://raw.githubusercontent.com/ory/kratos-selfservice-ui-react-native/master/src/components/Form/Form.tsx"
            />
          </TabItem>
          <TabItem value="input-form">
            <CodeFromRemote
              lang="handlebars"
              link="https://github.com/ory/kratos-selfservice-ui-react-native/blob/master/src/components/Form/Field.tsxs"
              src="https://raw.githubusercontent.com/ory/kratos-selfservice-ui-react-native/master/src/components/Form/Field.tsx"
            />
          </TabItem>
        </Tabs>
        The rest of the form partials can be found{' '}
        <a href="https://github.com/ory/kratos-selfservice-ui-react-native/tree/master/src">
          here
        </a>
      </TabItem>
    </Tabs>
  )
}

export default RenderFlow
