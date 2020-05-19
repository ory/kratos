So, you want to contribute to react-toggle? Great!

First off, setup:

  1. Fork the repo
  2. Clone it - `git clone https://github.com/{your-username-here}/react-toggle`
  3. Enter it - `cd react-toggle`
  4. Install the dependencies - `npm install`
  5. Run the development server - `npm run dev`
  6. Open `http://localhost:8080/webpack-dev-server/` in your web browser of choice

What we just did was start a [Webpack development server](https://webpack.github.io/docs/webpack-dev-server.html). Webpack is confusing, so I'm not going to go into how it works, but basically what it does is take the files from `src/`, bundles them up nicely, and serves them to us at the url above. One more nice thing it does is watch for changes, so when we edit files in `src/`, the webpage is automatically updated to reflect those changes, pretty neat, huh? (Note: The same site is available at `http://localhost:8080/`, but it won't be automatically updated. Make sure you're using the correct url or reloading manually)

Now it's time to make our changes. We usually just want to change files in the `src/` directory; it contains the source (hence the name) of our component and the documentation.

There's also the `spec/` directory, which contains the specification - or tests - for our component. If we add a new feature, we also want to add tests that make sure the feature works as it should and doesn't break anything.

The `docs/` directory contains what is shown at http://aaronshaf.github.io/react-toggle/ and should only be changed when publishing a new version (which @aaronshaf or one of the collaborators will take care of), so you don't need to worry about that.

So, in a simple list:

  1. Change some stuff in `src/`
  2. Check `http://localhost:8080/webpack-dev-server/` to see if it works as it should
  3. Run `npm test` to make sure we didn't break anything
  4. Add new tests to `spec/`
  5. Run `npm test` again, to make sure our new tests work
  6. Run `npm run lint` and fix all style errors (if there are any, if not, cudos)
  8. One last `npm test` again, because we edited stuff. (You can probably just skip the two previous ones. Should've read ahead ;))
  8. Add the relevant files - `git commit src/ spec/`
  9. Commit and write a nice, explanatory commit message.
  10. Push to your fork - `git push origin`
  11. Open a pull request

 You did it! :) Thank you!
