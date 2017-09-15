module Page.Helpers exposing (getFieldErrors, ifBelowLength, validClasses)

import Validate exposing (..)


-- FORM VALIDATION --


getFieldErrors : { b | field : field } -> List ( field, error ) -> List ( field, error )
getFieldErrors formField errors =
    let
        isFieldError error =
            let
                ( field, _ ) =
                    error
            in
                formField.field == field
    in
        List.filter isFieldError errors


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


isInvalid : List ( field, error ) -> { formField | dirty : Bool, field : field } -> Bool
isInvalid errors formField =
    formField.dirty && List.length (getFieldErrors formField errors) > 0


isValid : List ( field, error ) -> { formField | dirty : Bool, field : field } -> Bool
isValid errors formField =
    formField.dirty && List.length (getFieldErrors formField errors) == 0
