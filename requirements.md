so if a insurgency server is started and then my tacker app is started, it will not have the correct data

we have a few options for this:
    - dont do anything until the next mapload or maptravel event
    - dont do anything until the next server restart (lLogFileOpen event)
    - try to find the last mapload or maptravel event, and then read events from this point
        - to do this we will also need to ensure we dont process an event more than once