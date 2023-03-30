# RELEASE PROCESS

To create a new release (example is for release `v0.1.0`):

1. Increase the version according to Semantic Versioning in the [`VERSION` file](VERSION).
2. Run `make all` which updates `appVersion` in [the `Chart.yaml` of the helm chart.](charts/koor-operator/Chart.yaml) Update chart version.
3. Check out a new branch, which will be used for the pull request to update the version: `git checkout -b BRANCH_NAME`
4. Commit these changes now using `git commit -s -S`.
5. Push the branch using `git push -u origin BRANCH_NAME` with these changes and create a pull request on [GitHub](https://github.com/koor-tech/koor-operator).
6. Wait for pull request to be approved and merge it (if you have access to do so).
7. Create the new tag using `git tag v0.1.0` and then run `git push -u origin v0.1.0`
8. In a few minutes, the CI should have built and published a draft of the release here [GitHub - Releases List](https://github.com/koor-tech/koor-operator/releases).
9. Now edit the release and use the green button to publish the release.
10. Congratulations! The release is now fully published.
