# Scenario

**Feature**: the Command Code composite fetch becomes numeric values plus display strings

```
whoami + credits + subscription + summary -> ProjectUsageView -> UsageBlock
  7.5 of 10 credits left, 2.5 spent, 120 requests (5 failed)
  5-hour window 1.5/3, weekly window 1.5/6, 3 days to renewal
```

## Preconditions

- Inherits the fake Command Code API payload (`commandCodeBodies`) and the
  credentialed home.
- The subscription is `active` on plan `individual-go`, which carries a 10 credit
  monthly allowance named `Go`.

## Steps

1. Change nothing; assert the Command Code block.
