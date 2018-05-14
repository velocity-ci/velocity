module Page.Helpers
    exposing
        ( formatDate
        , formatTime
        , formatTimeSeconds
        , formatDateTime
        , sortByDatetime
        )

import Validate exposing (Validator, ifInvalid)
import Time.DateTime as DateTime exposing (DateTime)
import Time.Date as Date exposing (Date)


-- DATES --


formatDate : Date -> String
formatDate date =
    let
        ( year, month, day ) =
            Date.toTuple date
    in
        appendZero day ++ "/" ++ appendZero month ++ "/" ++ toString year


formatTime : DateTime -> String
formatTime dateTime =
    let
        ( _, _, _, hour, minute, _, _ ) =
            DateTime.toTuple dateTime
    in
        appendZero hour ++ ":" ++ appendZero minute


formatTimeSeconds : DateTime -> String
formatTimeSeconds dateTime =
    let
        ( _, _, _, hour, minute, second, _ ) =
            DateTime.toTuple dateTime
    in
        appendZero hour ++ ":" ++ appendZero minute ++ ":" ++ appendZero second


formatDateTime : DateTime -> String
formatDateTime dateTime =
    (formatDate (DateTime.date dateTime)) ++ " " ++ (formatTime dateTime)


sortByDatetime : (a -> DateTime) -> List a -> List a
sortByDatetime property items =
    items
        |> List.sortBy (property >> DateTime.toTimestamp)
        |> List.reverse



-- INTERNAL --


appendZero : Int -> String
appendZero int =
    if int < 10 then
        "0" ++ toString int
    else
        toString int
