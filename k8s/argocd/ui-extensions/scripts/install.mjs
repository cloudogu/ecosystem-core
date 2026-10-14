import {createHash} from "node:crypto";
import {execFileSync} from "node:child_process";
import {readFileSync} from "node:fs";

const [context, namespace = "argocd"] = process.argv.slice(2);
if (!context) {
  console.error("Usage: npm run install:cluster -- <kubernetes-context> [argocd-namespace]");
  process.exit(1);
}
const bundle = readFileSync(new URL("../dist/extension-ces-dogus.js", import.meta.url), "utf8");
const patch = JSON.parse(readFileSync(new URL("../install/argocd-server.patch.json", import.meta.url), "utf8"));
const base = ["--context", context, "--request-timeout=30s", "-n", namespace];
const kubectl = (args, input) => execFileSync("kubectl", [...base, ...args], {input, stdio: [input ? "pipe" : "inherit", "inherit", "inherit"]});

// Confirm that this is an existing ArgoCD installation before creating anything.
kubectl(["get", "deployment", "argocd-server", "-o", "name"]);
const configMap = {
  apiVersion: "v1",
  kind: "ConfigMap",
  metadata: {name: "argocd-ces-dogu-ui-extension", namespace},
  data: {"extension-ces-dogus.js": bundle},
};
patch.spec.template.metadata = {
  annotations: {"k8s.cloudogu.com/argocd-dogu-ui-sha256": createHash("sha256").update(bundle).digest("hex")},
};
kubectl(["apply", "-f", "-"], JSON.stringify(configMap));
kubectl(["patch", "deployment", "argocd-server", "--type=strategic", "-p", JSON.stringify(patch)]);
kubectl(["rollout", "status", "deployment/argocd-server", "--timeout=180s"]);
console.log("Installed. Reload ArgoCD, open a Blueprint resource, then select the Dogus tab.");
