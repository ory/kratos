import Tabs from '@theme/Tabs'
import TabItem from '@theme/TabItem'
import React from "react";
import CodeFromRemote from '../CodeFromRemote'

import registrationBrowser from './images/browser.png'

const RenderFlow = () => (
  <Tabs
    defaultValue="browser"
    values={[
      {label: 'Browser UI', value: 'browser'},
      {label: 'ExpressJS & Handlebars', value: 'express'},
      {label: 'ReactJS', value: 'react'},
      {label: 'React Native', value: 'react-native'},
    ]}>
    <TabItem value="browser">
      <img src={registrationBrowser} alt="User Registration HTML Form" />
    </TabItem>
    <TabItem value="express">
      <CodeFromRemote
        link="https://github.com/ory/kratos-selfservice-ui-node/blob/master/src/routes/login.ts"
        src="https://raw.githubusercontent.com/ory/kratos-selfservice-ui-node/master/src/routes/login.ts"
      />

      The views can be rather simple, as ORY Kratos provides you with all the
      information you need for rendering the forms.

      The following examples use Handlebars and a generic form generator to render the
      Flow:

      <Tabs
        defaultValue="registration"
        values={[
          {label: 'Registration View', value: 'registration'},
          {label: 'Generic Form View', value: 'generic-form'},
          {label: 'Example Input Form Element', value: 'input-form'},
        ]}>

        <TabItem value="registration">

          <CodeFromRemote
            link="https://github.com/ory/kratos-selfservice-ui-node/blob/master/views/registration.hbs"
            src="https://raw.githubusercontent.com/ory/kratos-selfservice-ui-node/master/views/registration.hbs"
          />

        </TabItem>

        <TabItem value="generic-form">

          <CodeFromRemote
            link="https://github.com/ory/kratos-selfservice-ui-node/blob/master/views/partials/form.hbs"
            src="https://raw.githubusercontent.com/ory/kratos-selfservice-ui-node/master/views/partials/form.hbs"
          />

        </TabItem>

        <TabItem value="input-form">

          <CodeFromRemote
            link="https://github.com/ory/kratos-selfservice-ui-node/blob/master/views/partials/form_input_default.hbs"
            src="https://raw.githubusercontent.com/ory/kratos-selfservice-ui-node/master/views/partials/form_input_default.hbs"
          />

        </TabItem>
      </Tabs>
      The rest of the form partials can be found{' '}
      <a href="https://github.com/ory/kratos-selfservice-ui-node/tree/master/views/partials">here</a>.
    </TabItem>
    <TabItem value="react">
      A React example is currently in the making.
    </TabItem>
    <TabItem value="react-native">
      A React Native example is currently in the making.
    </TabItem>
  </Tabs>
)

export default RenderFlow
