module Page.Helpers
    exposing
        ( formatDate
        , formatTime
        , formatTimeSeconds
        , formatDateTime
        , sortByDatetime
        , getFieldErrors
        , ifBelowLength
        , ifAboveLength
        , validClasses
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



-- FORM VALIDATION --


getFieldErrors : List ( field, error ) -> { b | field : field } -> List ( field, error )
getFieldErrors errors formField =
    List.filter (\e -> formField.field == Tuple.first e) errors


ifBelowLength : Int -> error -> Validator error String
ifBelowLength length =
    ifInvalid (\s -> String.length s < length)


ifAboveLength : Int -> error -> Validator error String
ifAboveLength length =
    ifInvalid (\s -> String.length s > length)


validClasses :
    List ( field, error )
    -> { formField | dirty : Bool, field : field }
    -> List ( String, Bool )
validClasses errors formField =
    [ ( "is-invalid", isInvalid errors formField )
    , ( "is-valid", isValid errors formField )
    ]



-- INTERNAL --


appendZero : Int -> String
appendZero int =
    if int < 10 then
        "0" ++ toString int
    else
        toString int


isInvalid : List ( field, error ) -> { formField | dirty : Bool, field : field } -> Bool
isInvalid errors formField =
    formField.dirty && List.length (getFieldErrors errors formField) > 0


isValid : List ( field, error ) -> { formField | dirty : Bool, field : field } -> Bool
isValid errors formField =
    formField.dirty && List.isEmpty (getFieldErrors errors formField)
