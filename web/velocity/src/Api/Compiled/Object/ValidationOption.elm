-- Do not manually edit this file, it was auto-generated by dillonkearns/elm-graphql
-- https://github.com/dillonkearns/elm-graphql


module Api.Compiled.Object.ValidationOption exposing (key, value)

import Api.Compiled.InputObject
import Api.Compiled.Interface
import Api.Compiled.Object
import Api.Compiled.Scalar
import Api.Compiled.ScalarCodecs
import Api.Compiled.Union
import Graphql.Internal.Builder.Argument as Argument exposing (Argument)
import Graphql.Internal.Builder.Object as Object
import Graphql.Internal.Encode as Encode exposing (Value)
import Graphql.Operation exposing (RootMutation, RootQuery, RootSubscription)
import Graphql.OptionalArgument exposing (OptionalArgument(..))
import Graphql.SelectionSet exposing (SelectionSet)
import Json.Decode as Decode


{-| The name of a variable to be subsituted in a validation message template
-}
key : SelectionSet String Api.Compiled.Object.ValidationOption
key =
    Object.selectionForField "String" "key" [] Decode.string


{-| The value of a variable to be substituted in a validation message template
-}
value : SelectionSet String Api.Compiled.Object.ValidationOption
value =
    Object.selectionForField "String" "value" [] Decode.string
