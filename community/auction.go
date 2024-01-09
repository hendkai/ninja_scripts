/* Created by Bull, Notriv and LordMike
 * It is not perfect. However, a quick start for many
 */

/*
 * VERSION 1.10
 */

/* DESCRIPTION
 * Automatically bid on auctions
 * Maximum resources customizable per type of auctioned item
 */

/*---------------------------------------------------------------------------------------------------------------------------*/

//######################################## SETTINGS START ########################################

// Add maximum bids here, the strings are substrings, so you could put 
// in 'metal' and bid that limit for all "metal" items
highestBids = {}
highestBids['bronze'] = 50000
highestBids['silver'] = 500000
highestBids['gold'] = 10000000
highestBids['platinum'] = 20000000

metBid = true                       // should you bid with metal?
crysBid = false                     // should you bid with crystal?
deutBid = false                     // should you bid with deuterium?
bidHome = "P:3:387:4"               // from which planet should be bid?

// Percentage of current resources to use as the maximum bid
maxBidPercentage = 0.1               // 10%

//######################################## SETTINGS END ########################################

var strings = import("strings")

ownPlayerID = GetCachedPlayer().PlayerID
celt = GetCachedCelestial(bidHome)
if celt == nil {
    LogError(bidHome + " is not one of your planet/moon")
    return
}

func DetermineMaxBid(name) {
    for key, highestBid in highestBids {
        if strings.Contains(name, key) {
            LogDebug("Detected '" + name + "' as '" + key + "', highest bid: " + Dotify(highestBid))
            return highestBid
        }
    }
    
    LogWarn("Unable to map '" + name + "' to a bid value, skipping")
    return 0
}

func CalculateMaxBid() {
    currentResources = GetCachedCelestial(bidHome).GetResources()
    maxBidPercentage = 0.1 // Setzen Sie den Prozentsatz neu (falls er sich geändert hat)

    maxBid = NewResources()
    maxBid.Metal = int64(float64(currentResources.Metal) * maxBidPercentage)
    maxBid.Crystal = int64(float64(currentResources.Crystal) * maxBidPercentage)
    maxBid.Deuterium = int64(float64(currentResources.Deuterium) * maxBidPercentage)

    LogDebug("Calculated max bid: " + Dotify(maxBid) + " (" + Dotify(maxBidPercentage * 100) + "% of current resources)")
}

func AucDo(ress) {
    bid = {}
    if metBid {
        bid = { celt.GetID() : NewResources(ress, 0, 0) }
    } else if crysBid {
        bid = { celt.GetID() : NewResources(0, ress, 0) }
    } else {
        bid = { celt.GetID() : NewResources(0, 0, ress) }
    }
    return DoAuction(bid)
}

func refreshTime(TimeEnd) {
    switch TimeEnd {     
        case TimeEnd <= 300:
            LogDebug("Only 5 min")
            return Random(2, 5)

        case TimeEnd <= 600:
            LogDebug("Only 10 Min")                        
            return Random(60, 120)

        case TimeEnd <= 1800:
            LogDebug("Only 30 Min")                        
            return Random(180, 300)

        case TimeEnd <= 3600:
            LogDebug("Only 60 Min")                        
            return Random(300, 600)

        default:
            LogError("Unknown TimeEnd value", TimeEnd)
            return Random(15, 30)  // Erhöhen Sie die Wartezeit auf 15 bis 30 Sekunden
    }
}

func customSleep(sleepTime) {
    if sleepTime <= 0 {
        sleepTime = Random(5, 10)
    }
    LogDebug("Wait " + ShortDur(sleepTime))
    Sleep(sleepTime * 1000)
}

func didWon(auc) {
    if auc.HighestBidderUserID == ownPlayerID {
        LogInfo("You won the auction with " + Dotify(auc.CurrentBid) + " resources!")
    }
}

func processAuction() {
    auc, err = GetAuction()
    if err != nil {
        LogError(err)
        return Random(5, 10)
    }
    
    // Log-Ausgabe für Debugging der Auktionsdaten
    LogDebug("Auction data: " + auc)

    if auc.HasFinished {
        if auc.Endtime > 7200 {
            LogInfo("There is currently no auction")
        } else {
            LogInfo("Auction has finished")
        }
        didWon(auc)
        return auc.Endtime + 10
    }
    
    highestBid = DetermineMaxBid(auc.CurrentItem)
    if highestBid <= 0 {
        // Wenn kein fester Höchstgebot gefunden wird, berechne den maximalen Gebotsbetrag
        CalculateMaxBid()
        LogInfo("No fixed bid found for '" + auc.CurrentItem + "', using calculated max bid")
        highestBid = maxBid.Total()
    }
    
    if auc.AlreadyBid == 0 {
        LogDebug("Willing to bid " + Dotify(highestBid) + " for '" + auc.CurrentItem + "'")
    }
    
    if auc.HighestBidderUserID == ownPlayerID {
        LogDebug("Already highest bidder for '" + auc.CurrentItem + "' at " + Dotify(auc.CurrentBid) + " / " + Dotify(highestBid) + ", waiting..")
        return refreshTime(auc.Endtime)
    }
    if auc.MinimumBid > highestBid {
        LogWarn("Resources exceeded for '" + auc.CurrentItem + "', currently at " + Dotify(auc.CurrentBid) + " / " + Dotify(highestBid) + "")
        return auc.Endtime + 10
    }

    shouldBid = auc.MinimumBid - auc.AlreadyBid
    LogInfo("Bidding " + Dotify(auc.AlreadyBid) + " + " + Dotify(shouldBid) + " / " + Dotify(highestBid) + " resources for '" + auc.CurrentItem + "'")
    err = AucDo(shouldBid)
    if err != nil {
        LogError(err)
        return Random(5, 10)
    }
    return refreshTime(auc.Endtime)
}

func doWork() {
    for { // forever process auctions
        sleepTime = processAuction()
        customSleep(sleepTime)
    }
}
doWork()
