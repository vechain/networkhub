import * as core from '@actions/core';
import * as tc from '@actions/tool-cache';
import * as github from '@actions/github';
import * as process from 'process';
import * as fs from 'fs';
import * as path from 'path';

function getExecutableName(): string {
    let platform: string

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

    let arch: string

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

    return `network-hub-${platform}-${arch}${process.platform === 'win32' ? '.exe' : ''}`
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

    // set up auth/environment
    const token = core.getInput("token");
    if (!token) {
        throw new Error(
            `No GitHub token found`
        )
    }
    const octokit = github.getOctokit(token)
    const release = await octokit.rest.repos.getReleaseByTag({
        tag: version,
        owner: 'vechain',
        repo: 'networkhub'
    })

    const executableName = getExecutableName()

    const asset = release.data.assets.find(asset => {
        return asset.name === executableName
    })

    if (!asset) {
        throw new Error(`No asset found for ${executableName}`)
    }

    const destination = path.join(__dirname, 'network-hub')

    core.info(`Downloading network-hub from ${asset.url}`)
    const binPath = await tc.downloadTool(asset.url,
      destination,
      `token ${token}`,
      {
          accept: 'application/octet-stream'
      }
    );

    // list the files in the binPath
    fs.readdirSync(__dirname).forEach(file => {
        core.info(file);
    });

    core.info(`Successfully downloaded network-hub to ${binPath}`)

    fs.chmodSync(binPath, '755');
    //
    // let extractArgs = core.getMultilineInput("extractArgs");
    // let extractedPath = await tc.extractTar(binPath, undefined, extractArgs);
    // core.info(`Successfully extracted network-hub to ${extractedPath}`)
    core.addPath(destination);
}

setup().catch((error) => {
    core.error(error)
    core.setFailed(error);
});
