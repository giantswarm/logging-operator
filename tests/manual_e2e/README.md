# Manual e2e testing

As of now, the [apptest-framework](https://github.com/giantswarm/apptest-framework) used for automated e2e testiong doesn't support MC-only apps. Hence the manual procedure described here in order to ensure that the app works as expected in a Giant Swarm environment.

## Procedure

Before proceeding to any kind of test, you'll first have to deploy your custom branch app's version into a testing installation. Don't forget to suspend flux reconciliation for this app during the whole testing process. See [here](https://intranet.giantswarm.io/docs/dev-and-releng/flux/suspending-flux/#how-to-be-more-granular--subtle-with-suspending-resources-and-why-be-careful-with-this) for details on how to evict an app from flux's reconciliation.

Then, run the `tests/manual_e2e/basic_test.sh <installation>` command from the repo's root. This will create a WC and make sure the logs collector is deployed and configured correctly.

Once that's done, there are additional things you should do :

- Wait for ~ 10min after running the above command.
- head over to the "Explore" panel from the installation's grafana and make sure that you can see logs from several WCs, especially the`loggingoperatortest` one using any logql command.
- Check for new alerts on the `Alerts timeline` dashboard.
- If everything appears to be fine, then you can revert the flux's evicting procedure that you did and let it reconcile to its original version.

Congratulations, you have completed the manual e2e testing procedure ! Your PR is now ready to be merged.
