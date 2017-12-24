module Views.Build exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Build as Build exposing (Build)
import Data.BuildStep as BuildStep exposing (BuildStep)


viewBuildContainer : Build -> List BuildStep -> Html msg
viewBuildContainer build steps =
    div []
        [ h3 [] [ text ("build container " ++ Build.idToString build.id) ]
        , div [] (List.map viewBuildStep steps)
        ]


viewBuildStep : BuildStep -> Html msg
viewBuildStep step =
    div []
        [ text (BuildStep.idToString step.id) ]
