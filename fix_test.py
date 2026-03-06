import re

with open('pkg/bgpengine/engine_test.go', 'r') as f:
    content = f.read()

# Replace the test to account for the stats worker running asynchronously
# e.prefixToClassification and e.currentAnomalies are no longer public fields updated immediately
# They are local to the stats_worker. So this test doesn't make sense anymore if those fields aren't public.
