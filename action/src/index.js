const core = require('@actions/core');
const tc = require('@actions/tool-cache');

/**
 * Get the download URL for the tool to be installed
 * @param {string} version - The version of the tool to be installed
 * @returns {Promise<string>}
 */
async function getDownloadURL(version) {
    // Get the platform (i.e. linux, darwin, win32)
    const platform = process.platform;
    // Get the architecture (i.e. x64, arm64)
    const arch = process.arch;

    console.log(`Platform: ${platform}, Arch: ${arch}`);

    const url = `https://github.com/vechain/networkhub/releases/download/${version}/networkHub-${platform}-${arch}${platform === 'win32' ? '.exe' : ''}`;

    console.log(`Download URL: ${url}`);

    return url;
}

async function setup() {
    // Get version of tool to be installed
    const version = core.getInput('version');
    if (!tc.isExplicitVersion(version)) {
        core.setFailed('No version specified')
        return
    }

    // Download the specific version of the tool, e.g. as a tarball
    const pathToTarball = await tc.downloadTool(await getDownloadURL(version));

    // Extract the tarball onto the runner
    const pathToCLI = await tc.extractTar(pathToTarball);

    // Expose the tool by adding it to the PATH
    core.addPath(pathToCLI)
}

module.exports = setup
