'use strict';
const {promisify} = require('util');
const path = require('path');
const childProcess = require('child_process');
const fs = require('fs');
const isWsl = require('is-wsl');
const isDocker = require('is-docker');

const pAccess = promisify(fs.access);
const pExecFile = promisify(childProcess.execFile);

// Path to included `xdg-open`.
const localXdgOpenPath = path.join(__dirname, 'xdg-open');

// Convert a path from WSL format to Windows format:
// `/mnt/c/Program Files/Example/MyApp.exe` → `C:\Program Files\Example\MyApp.exe`
const wslToWindowsPath = async path => {
	const {stdout} = await pExecFile('wslpath', ['-w', path]);
	return stdout.trim();
};

module.exports = async (target, options) => {
	if (typeof target !== 'string') {
		throw new TypeError('Expected a `target`');
	}

	options = {
		wait: false,
		background: false,
		url: false,
		...options
	};

	let command;
	let appArguments = [];
	const cliArguments = [];
	const childProcessOptions = {};

	if (Array.isArray(options.app)) {
		appArguments = options.app.slice(1);
		options.app = options.app[0];
	}

	// Encodes the target as if it were an URL. Especially useful to get
	// double-quotes through the “double-quotes on Windows caveat”, but it
	// can be used on any platform.
	if (options.url) {
		target = encodeURI(target);

		if (isWsl) {
			target = target.replace(/&/g, '^&');
		}
	}

	if (process.platform === 'darwin') {
		command = 'open';

		if (options.wait) {
			cliArguments.push('--wait-apps');
		}

		if (options.background) {
			cliArguments.push('--background');
		}

		if (options.app) {
			cliArguments.push('-a', options.app);
		}
	} else if (process.platform === 'win32' || (isWsl && !isDocker())) {
		command = 'cmd' + (isWsl ? '.exe' : '');
		cliArguments.push('/s', '/c', 'start', '""', '/b');

		if (!isWsl) {
			// Always quoting target allows for URLs/paths to have spaces and unmarked characters, as `cmd.exe` will
			// interpret them as plain text to be forwarded as one unique argument. Enabling `windowsVerbatimArguments`
			// disables Node.js's default quotes and escapes handling (https://git.io/fjdem).
			// References:
			// - Issues #17, #44, #55, #77, #101, #115
			// - Pull requests: #74, #98
			//
			// As a result, all double-quotes are stripped from the `target` and do not get to your desired destination.
			target = `"${target}"`;
			childProcessOptions.windowsVerbatimArguments = true;

			if (options.app) {
				options.app = `"${options.app}"`;
			}
		}

		if (options.wait) {
			cliArguments.push('/wait');
		}

		if (options.app) {
			if (isWsl && options.app.startsWith('/mnt/')) {
				const windowsPath = await wslToWindowsPath(options.app);
				options.app = windowsPath;
			}

			cliArguments.push(options.app);
		}

		if (appArguments.length > 0) {
			cliArguments.push(...appArguments);
		}
	} else {
		if (options.app) {
			command = options.app;
		} else {
			// When bundled by Webpack, there's no actual package file path and no local `xdg-open`.
			const isBundled = !__dirname || __dirname === '/';

			// Check if local `xdg-open` exists and is executable.
			let exeLocalXdgOpen = false;
			try {
				await pAccess(localXdgOpenPath, fs.constants.X_OK);
				exeLocalXdgOpen = true;
			} catch (_) {}

			const useSystemXdgOpen = process.versions.electron ||
				process.platform === 'android' || isBundled || !exeLocalXdgOpen;
			command = useSystemXdgOpen ? 'xdg-open' : localXdgOpenPath;
		}

		if (appArguments.length > 0) {
			cliArguments.push(...appArguments);
		}

		if (!options.wait) {
			// `xdg-open` will block the process unless stdio is ignored
			// and it's detached from the parent even if it's unref'd.
			childProcessOptions.stdio = 'ignore';
			childProcessOptions.detached = true;
		}
	}

	cliArguments.push(target);

	if (process.platform === 'darwin' && appArguments.length > 0) {
		cliArguments.push('--args', ...appArguments);
	}

	const subprocess = childProcess.spawn(command, cliArguments, childProcessOptions);

	if (options.wait) {
		return new Promise((resolve, reject) => {
			subprocess.once('error', reject);

			subprocess.once('close', exitCode => {
				if (exitCode > 0) {
					reject(new Error(`Exited with code ${exitCode}`));
					return;
				}

				resolve(subprocess);
			});
		});
	}

	subprocess.unref();

	return subprocess;
};
