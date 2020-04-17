export const handleConnectStatus = (setConnectStatus) => (status) => {
    const update = {};
    update.lastUpdate = new Date(status.DataSnapshotTime);
    status.MetricResults.forEach(r => r.Collections.forEach(({ Metric: { Name: n }, Value: v }) => {
        switch (n) {
            case 'AGENTS_ONLINE':
                update.agentsOnline = v
                break;
            case 'AGENTS_AVAILABLE':
                update.agentsAvailable = v;
                break;
            case 'AGENTS_ON_CALL':
                update.agentsOnCall = v;
                break;
            case 'CONTACTS_IN_QUEUE':
                update.contactsInQueue = v;
                break;
            default:
                console.log(`Ignored metric ${n} with value ${v}`);
        }
    }));

    setConnectStatus(update);
}
