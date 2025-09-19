#!groovy
@Library('github.com/cloudogu/ces-build-lib@4.3.0')
import com.cloudogu.ces.cesbuildlib.*

git = new Git(this, "cesmarvin")
git.committerName = 'cesmarvin'
git.committerEmail = 'cesmarvin@cloudogu.com'
gitflow = new GitFlow(this, git)
github = new GitHub(this, git)
changelog = new Changelog(this)
makefile = new Makefile(this)
Docker docker = new Docker(this)

repositoryName = "ecosystem-core"
productionReleaseBranch = "main"
developmentBranch = "develop"
currentBranch = "${env.BRANCH_NAME}"

registryNamespace = "k8s"
registryUrl = "registry.cloudogu.com"

goVersion = "1.25.1"
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

                stage('Lint Dockerfile') {
                    Dockerfile dockerfile = new Dockerfile(this)
                    dockerfile.lint("./default-config/Dockerfile")
                }

                stage('Check markdown links') {
                    Markdown markdown = new Markdown(this, "3.11.0")
                    markdown.check()
                }

                docker
                        .image("golang:${goVersion}")
                        .mountJenkinsUser()
                        .inside("--volume ${WORKSPACE}:/${repositoryName} -w /${repositoryName}")
                                {
                                    stage('Build & Unit-Test') {
                                        sh "make test-default-config"
                                        junit allowEmptyResults: true, testResults: 'default-config/target/*-tests.xml'
                                    }

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

                stage('SonarQube') {
                    stageStaticAnalysisSonarQube()
                }

                K3d k3d = new K3d(this, "${WORKSPACE}", "${WORKSPACE}/k3d", env.PATH)

                try {
                    stage('Set up k3d cluster') {
                        k3d.startK3d()
                    }

                    String version = makefile.getVersion()
                    def imageNameDefaultConfig = ""
                    stage('Build & Push Image') {
						imageNameDefaultConfig = buildAndPushToLocalRegistry(k3d, "cloudogu/${repositoryName}-default-config", version, "./default-config")
                    }

                    stage('Update development resources') {
                    	def repository = imageNameDefaultConfig.substring(0, imageNameDefaultConfig.lastIndexOf(":"))
                        docker.image("golang:${goVersion}")
                        	.mountJenkinsUser()
                            .inside("--volume ${WORKSPACE}:/workdir -w /workdir") {
                            	sh "STAGE=development IMAGE_DEV=${repository} make helm-values-replace-image-repo"
                        	}
                    }

                    stage('Deploy ecosystem-core') {
                        withCredentials([[$class: 'UsernamePasswordMultiBinding', credentialsId: 'harborhelmchartpush', usernameVariable: 'HARBOR_USERNAME', passwordVariable: 'HARBOR_PASSWORD']]) {
                            k3d.helm("registry login ${registryUrl} --username '${HARBOR_USERNAME}' --password '${HARBOR_PASSWORD}'")
                            k3d.helm("install k8s-component-operator-crd oci://registry.cloudogu.com/k8s/k8s-component-operator-crd  --version 1.10.0")
                            k3d.helm("registry logout ${registryUrl}")

                            k3d.helm("install ${repositoryName} ${helmChartDir}")
                        }
                    }

                    stage('Test ecosystem-core') {
                        k3d.kubectl("wait --for=condition=ready pod -l app.kubernetes.io/name=k8s-component-operator --timeout=300s")
                    }
                } catch (Exception e) {
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

void gitWithCredentials(String command) {
    withCredentials([usernamePassword(credentialsId: 'cesmarvin', usernameVariable: 'GIT_AUTH_USR', passwordVariable: 'GIT_AUTH_PSW')]) {
        sh(
                script: "git -c credential.helper=\"!f() { echo username='\$GIT_AUTH_USR'; echo password='\$GIT_AUTH_PSW'; }; f\" " + command,
                returnStdout: true
        )
    }
}

void stageStaticAnalysisSonarQube() {
    def scannerHome = tool name: 'sonar-scanner', type: 'hudson.plugins.sonar.SonarRunnerInstallation'
    withSonarQubeEnv {
        sh "git config 'remote.origin.fetch' '+refs/heads/*:refs/remotes/origin/*'"
        gitWithCredentials("fetch --all")

        if (currentBranch == productionReleaseBranch) {
            echo "This branch has been detected as the production branch."
            sh "${scannerHome}/bin/sonar-scanner -Dsonar.branch.name=${env.BRANCH_NAME}"
        } else if (currentBranch == developmentBranch) {
            echo "This branch has been detected as the development branch."
            sh "${scannerHome}/bin/sonar-scanner -Dsonar.branch.name=${env.BRANCH_NAME}"
        } else if (env.CHANGE_TARGET) {
            echo "This branch has been detected as a pull request."
            sh "${scannerHome}/bin/sonar-scanner -Dsonar.pullrequest.key=${env.CHANGE_ID} -Dsonar.pullrequest.branch=${env.CHANGE_BRANCH} -Dsonar.pullrequest.base=${developmentBranch}"
        } else if (currentBranch.startsWith("feature/")) {
            echo "This branch has been detected as a feature branch."
            sh "${scannerHome}/bin/sonar-scanner -Dsonar.branch.name=${env.BRANCH_NAME}"
        } else {
            echo "This branch has been detected as a miscellaneous branch."
            sh "${scannerHome}/bin/sonar-scanner -Dsonar.branch.name=${env.BRANCH_NAME} "
        }
    }
    timeout(time: 2, unit: 'MINUTES') { // Needed when there is no webhook for example
        def qGate = waitForQualityGate()
        if (qGate.status != 'OK') {
            unstable("Pipeline unstable due to SonarQube quality gate failure")
        }
    }
}

void stageAutomaticRelease() {
    if (gitflow.isReleaseBranch()) {
        String releaseVersion = makefile.getVersion()
        String changelogVersion = git.getSimpleBranchName()

        stage('Push Helm chart to Harbor') {
            docker
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

def buildAndPushToLocalRegistry(def k3d, def imageName, def tag, def dockerFile) {
    def internalHandle="${imageName}:${tag}"
    def externalRegistry="${k3d.@registry.@imageRegistryExternalHandle}"

    def dockerImage = this.docker.build("${internalHandle}", "${dockerFile}")

    this.docker.withRegistry("http://${externalRegistry}/") {
        dockerImage.push("${tag}")
    }

    return "${k3d.@registry.@imageRegistryInternalHandle}/${internalHandle}"
}