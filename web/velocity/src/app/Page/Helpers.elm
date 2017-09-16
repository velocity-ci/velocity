module Page.Helpers
    exposing
        ( formatDate
        , formatTime
        , formatDateTime
        , sortByDatetime
        , getFieldErrors
        , ifBelowLength
        , validClasses
        )

import Validate exposing (Validator, ifInvalid)
import Time.DateTime as DateTime exposing (DateTime)


-- DATES --


formatDate : DateTime -> String
formatDate dateTime =
    let
        ( year, month, day, _, _, _, _ ) =
            DateTime.toTuple dateTime
    in
        appendZero day ++ "/" ++ appendZero month ++ "/" ++ toString year


formatTime : DateTime -> String
formatTime dateTime =
    let
        ( _, _, _, hour, minute, _, _ ) =
            DateTime.toTuple dateTime
    in
        appendZero hour ++ ":" ++ appendZero minute


formatDateTime : DateTime -> String
formatDateTime dateTime =
    (formatDate dateTime) ++ ":" ++ (formatTime dateTime)


sortByDatetime : (a -> DateTime) -> List a -> List a
sortByDatetime property items =
    items
        |> List.sortBy (property >> DateTime.toTimestamp)
        |> List.reverse



-- FORM VALIDATION --


getFieldErrors : { b | field : field } -> List ( field, error ) -> List ( field, error )
getFieldErrors formField errors =
    List.filter (\e -> formField.field == Tuple.first e) errors


ifBelowLength : Int -> error -> Validator error String
ifBelowLength length =
    ifInvalid (\s -> String.length s < length)


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
    formField.dirty && List.length (getFieldErrors formField errors) > 0


isValid : List ( field, error ) -> { formField | dirty : Bool, field : field } -> Bool
isValid errors formField =
    formField.dirty && List.length (getFieldErrors formField errors) == 0
