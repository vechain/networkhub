const core = require('@actions/core');
const tc = require('@actions/tool-cache');

/**
 * Get the download URL for the tool to be installed
 * @param {string} version - The version of the tool to be installed
 * @returns {Promise<string>}
 */
async function getDownloadURL(version) {
    let platform;

    switch (process.platform) {
        case 'win32':
            platform = 'windows';
            break;
        case 'darwin':
            platform = 'macos';
            break;
        case 'linux':
            platform = 'linux';
            break;
        default:
            throw new Error(`Unsupported platform: ${process.platform}`);
    }

    let arch;

    switch (process.arch) {
        case 'x64':
            arch = 'amd64';
            break;
        case 'arm64':
            arch = 'arm64';
            break;
        default:
            throw new Error(`Unsupported architecture: ${process.arch}`);
    }

    core.info(`Platform: ${platform}`)
    core.info(`Arch: ${arch}`)

    // https://github.com/vechain/networkhub/releases/download/v0.0.3/network-hub-macos-arm64

    // https://github.com/vechain/networkhub/releases/download/v0.0.3/network-hub-macos-arm64


    const url = `https://github.com/vechain/networkhub/releases/download/${version}/network-hub-${platform}-${arch}${process.platform === 'win32' ? '.exe' : ''}`;
    core.info(`Download URL: ${url}`)
    return url;
}

async function setup() {
    // Get version of tool to be installed
    const version = core.getInput('version');
    if (!tc.isExplicitVersion(version)) {
        core.setFailed('No version specified')
        return
    }

    core.setOutput('version', version)

    core.info(`Installing networkHub version ${version}`)

    //create an auth header using the token provided
    const token = core.getInput('token');
    if (!token) {
        core.setFailed('No token specified')
        return
    }
    // Download the specific version of the tool, e.g. as a tarball
    const pathToCLI = await tc.downloadTool(await getDownloadURL(version), undefined, `token ${token}`);

    // Expose the tool by adding it to the PATH
    core.addPath(pathToCLI)
}

setup().catch((error) => {
    core.error(error)
    core.setFailed(error);
});
