module Component.Form
    exposing
        ( Context
        , FormField
        , Error
        , updateInput
        , newField
        , submitting
        , submit
        , optionalError
        , allErrors
        , globalErrors
        , updateServerErrors
        , resetServerErrorsForField
        , ifBelowLength
        , ifAboveLength
        , validClasses
        , getFieldErrors
        )

import Request.Errors
import Json.Decode as Decode exposing (Decoder, string)
import Json.Decode.Pipeline as Pipeline exposing (optional)
import Validate exposing (Validator, ifInvalid)


type alias FormField f =
    { value : String
    , dirty : Bool
    , field : f
    }


type alias Error f =
    ( f, String )


type alias Context field form =
    { form : form
    , errors : List (Error field)
    , serverErrors : List (Error field)
    , submitting : Bool
    }



-- UPDATE HELPERS --


newField : f -> FormField f
newField field =
    FormField "" False field


updateInput : f -> String -> FormField f
updateInput field value =
    FormField value True field


updateServerErrors : List ( String, String ) -> (( String, String ) -> Error f) -> Context f x -> Context f x
updateServerErrors errorMessages serverErrorToFormError context =
    { context | serverErrors = List.map serverErrorToFormError errorMessages }


resetServerErrorsForField : Context field form -> field -> List (Error field)
resetServerErrorsForField context field =
    resetServerErrors context.serverErrors field


submit : Context field form -> Context field form
submit context =
    { context
        | submitting = True
        , serverErrors = []
        , errors = []
    }


submitting : Bool -> Context field form -> Context field form
submitting submitting context =
    { context | submitting = submitting }


optionalError : String -> Decoder (List ( String, String ) -> a) -> Decoder a
optionalError fieldName =
    let
        errorToTuple errorMessage =
            ( fieldName, errorMessage )
    in
        optional fieldName (Decode.list (Decode.map errorToTuple string)) []


allErrors : Context field msg -> List (Error field)
allErrors { errors, serverErrors } =
    errors ++ serverErrors


globalErrors : field -> List (Error field) -> List (Error field)
globalErrors globalField errors =
    List.filter (\e -> (Tuple.first e) == globalField) errors



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


isInvalid : List ( field, error ) -> { formField | dirty : Bool, field : field } -> Bool
isInvalid errors formField =
    formField.dirty && List.length (getFieldErrors errors formField) > 0


isValid : List ( field, error ) -> { formField | dirty : Bool, field : field } -> Bool
isValid errors formField =
    formField.dirty && List.isEmpty (getFieldErrors errors formField)



-- PRIVATE --


resetServerErrors : List (Error f) -> f -> List (Error f)
resetServerErrors errors field =
    let
        shouldInclude error =
            Tuple.first error /= field
    in
        List.filter shouldInclude errors
