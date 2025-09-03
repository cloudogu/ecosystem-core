#!groovy
@Library('github.com/cloudogu/ces-build-lib@4.3.0')
import com.cloudogu.ces.cesbuildlib.*

git = new Git(this, "cesmarvin")
git.committerName = 'cesmarvin'
git.committerEmail = 'cesmarvin@cloudogu.com'
gitflow = new GitFlow(this, git)
github = new GitHub(this, git)
changelog = new Changelog(this)

repositoryName = "ecosystem-core"
productionReleaseBranch = "main"

registryNamespace = "k8s"
registryUrl = "registry.cloudogu.com"

goVersion = "1.24"
helmTargetDir = "target/k8s"
helmChartDir = "${helmTargetDir}/helm"

node('docker') {
    timestamps {
        catchError {
            timeout(activity: false, time: 60, unit: 'MINUTES') {
                stage('Checkout') {
                    checkout scm
                    make 'clean'
                }

                new Docker(this)
                        .image("golang:${goVersion}")
                        .mountJenkinsUser()
                        .inside("--volume ${WORKSPACE}:/${repositoryName} -w /${repositoryName}")
                               {
                               		helmRegistryLogin()

                                    stage('Generate k8s Resources') {
                                        make 'helm-update-dependencies'
                                        make 'helm-generate'
                                        archiveArtifacts "${helmTargetDir}/**/*"
                                    }

                                    stage("Lint helm") {
                                        make 'helm-lint'
                                    }
                                }

                K3d k3d = new K3d(this, "${WORKSPACE}", "${WORKSPACE}/k3d", env.PATH)

                try {
                    stage('Set up k3d cluster') {
                        k3d.startK3d()
                    }

                    stage('Deploy ecosystem-core') {
                        k3d.helm("install ${repositoryName} ${helmChartDir}")
                    }

                    stage('Test ecosystem-core') {
                        k3d.kubectl("wait --for=condition=ready pod -l app.kubernetes.io/instance=k8s-component-operator --timeout=300s")
                    }
                } catch(Exception e) {
                    k3d.collectAndArchiveLogs()
                    throw e as java.lang.Throwable
                } finally {
                    stage('Remove k3d cluster') {
                        k3d.deleteK3d()
                    }
                }
            }
        }

        stageAutomaticRelease()
    }
}

void stageAutomaticRelease() {
    if (gitflow.isReleaseBranch()) {
        Makefile makefile = new Makefile(this)
        String releaseVersion = makefile.getVersion()
        String changelogVersion = git.getSimpleBranchName()

        stage('Push Helm chart to Harbor') {
            new Docker(this)
                    .image("golang:${goVersion}")
                    .mountJenkinsUser()
                    .inside("--volume ${WORKSPACE}:/${repositoryName} -w /${repositoryName}")
                            {
                                make 'helm-package'
                                archiveArtifacts "${helmTargetDir}/**/*"

                                withCredentials([[$class: 'UsernamePasswordMultiBinding', credentialsId: 'harborhelmchartpush', usernameVariable: 'HARBOR_USERNAME', passwordVariable: 'HARBOR_PASSWORD']]) {
                                    sh ".bin/helm registry login ${registryUrl} --username '${HARBOR_USERNAME}' --password '${HARBOR_PASSWORD}'"
                                    sh ".bin/helm push ${helmChartDir}/${repositoryName}-${releaseVersion}.tgz oci://${registryUrl}/${registryNamespace}"
                                }
                            }
        }

        stage('Finish Release') {
            gitflow.finishRelease(changelogVersion, productionReleaseBranch)
        }

        stage('Add Github-Release') {
            releaseId = github.createReleaseWithChangelog(changelogVersion, changelog, productionReleaseBranch)
        }
    }
}

void make(String makeArgs) {
    sh "make ${makeArgs}"
}

void helmRegistryLogin() {
    make 'install-helm'

	withCredentials([[$class: 'UsernamePasswordMultiBinding', credentialsId: 'harborhelmchartpush', usernameVariable: 'HARBOR_USERNAME', passwordVariable: 'HARBOR_PASSWORD']]) {
    	sh ".bin/helm registry login ${registryUrl} --username '${HARBOR_USERNAME}' --password '${HARBOR_PASSWORD}'"
    }
}