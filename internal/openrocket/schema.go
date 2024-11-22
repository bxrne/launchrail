// Package openrocket provides the data structures for OpenRocket file format.
// INFO: The schema represents OpenRocket XML format version 1.0
// TODO: Implement version detection and multiple schema support
// WARN: This schema is auto-generated and may need manual updates for new OpenRocket versions
package openrocket

import (
	"encoding/xml"
)

// TODO: Do the below a better way that will handle OpenRocket updates or else do versions?
// Openrocket was generated 2024-10-18 00:32:31 by https://xml-to-go.github.io/ in Ukraine.
type Openrocket struct {
	XMLName xml.Name `xml:"openrocket"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Creator string   `xml:"creator,attr"`
	Rocket  struct {
		Text        string `xml:",chardata"`
		Name        string `xml:"name"`
		ID          string `xml:"id"`
		Axialoffset struct {
			Text   string `xml:",chardata"`
			Method string `xml:"method,attr"`
		} `xml:"axialoffset"`
		Position struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"position"`
		Comment            string `xml:"comment"`
		Designer           string `xml:"designer"`
		Revision           string `xml:"revision"`
		Motorconfiguration []struct {
			Text     string `xml:",chardata"`
			Configid string `xml:"configid,attr"`
			Default  string `xml:"default,attr"`
			Stage    struct {
				Text   string `xml:",chardata"`
				Number string `xml:"number,attr"`
				Active string `xml:"active,attr"`
			} `xml:"stage"`
		} `xml:"motorconfiguration"`
		Referencetype string `xml:"referencetype"`
		Subcomponents struct {
			Text  string `xml:",chardata"`
			Stage struct {
				Text          string `xml:",chardata"`
				Name          string `xml:"name"`
				ID            string `xml:"id"`
				Subcomponents struct {
					Text     string `xml:",chardata"`
					Nosecone struct {
						Text       string `xml:",chardata"`
						Name       string `xml:"name"`
						ID         string `xml:"id"`
						Appearance struct {
							Text  string `xml:",chardata"`
							Paint struct {
								Text  string `xml:",chardata"`
								Red   string `xml:"red,attr"`
								Green string `xml:"green,attr"`
								Blue  string `xml:"blue,attr"`
								Alpha string `xml:"alpha,attr"`
							} `xml:"paint"`
							Shine string `xml:"shine"`
						} `xml:"appearance"`
						Finish   string `xml:"finish"`
						Material struct {
							Text    string `xml:",chardata"`
							Type    string `xml:"type,attr"`
							Density string `xml:"density,attr"`
						} `xml:"material"`
						Length               string `xml:"length"`
						Thickness            string `xml:"thickness"`
						Shape                string `xml:"shape"`
						Shapeparameter       string `xml:"shapeparameter"`
						Aftradius            string `xml:"aftradius"`
						Aftshoulderradius    string `xml:"aftshoulderradius"`
						Aftshoulderlength    string `xml:"aftshoulderlength"`
						Aftshoulderthickness string `xml:"aftshoulderthickness"`
						Aftshouldercapped    string `xml:"aftshouldercapped"`
						Isflipped            string `xml:"isflipped"`
						Subcomponents        struct {
							Text      string `xml:",chardata"`
							Parachute []struct {
								Text        string `xml:",chardata"`
								Name        string `xml:"name"`
								ID          string `xml:"id"`
								Axialoffset struct {
									Text   string `xml:",chardata"`
									Method string `xml:"method,attr"`
								} `xml:"axialoffset"`
								Position struct {
									Text string `xml:",chardata"`
									Type string `xml:"type,attr"`
								} `xml:"position"`
								Packedlength    string `xml:"packedlength"`
								Packedradius    string `xml:"packedradius"`
								Radialposition  string `xml:"radialposition"`
								Radialdirection string `xml:"radialdirection"`
								Cd              string `xml:"cd"`
								Material        struct {
									Text    string `xml:",chardata"`
									Type    string `xml:"type,attr"`
									Density string `xml:"density,attr"`
								} `xml:"material"`
								Deployevent    string `xml:"deployevent"`
								Deployaltitude string `xml:"deployaltitude"`
								Deploydelay    string `xml:"deploydelay"`
								Diameter       string `xml:"diameter"`
								Linecount      string `xml:"linecount"`
								Linelength     string `xml:"linelength"`
								Linematerial   struct {
									Text    string `xml:",chardata"`
									Type    string `xml:"type,attr"`
									Density string `xml:"density,attr"`
								} `xml:"linematerial"`
							} `xml:"parachute"`
						} `xml:"subcomponents"`
					} `xml:"nosecone"`
					Bodytube struct {
						Text       string `xml:",chardata"`
						Name       string `xml:"name"`
						ID         string `xml:"id"`
						Appearance struct {
							Text  string `xml:",chardata"`
							Paint struct {
								Text  string `xml:",chardata"`
								Red   string `xml:"red,attr"`
								Green string `xml:"green,attr"`
								Blue  string `xml:"blue,attr"`
								Alpha string `xml:"alpha,attr"`
							} `xml:"paint"`
							Shine string `xml:"shine"`
						} `xml:"appearance"`
						Finish   string `xml:"finish"`
						Material struct {
							Text    string `xml:",chardata"`
							Type    string `xml:"type,attr"`
							Density string `xml:"density,attr"`
						} `xml:"material"`
						Length        string `xml:"length"`
						Thickness     string `xml:"thickness"`
						Radius        string `xml:"radius"`
						Subcomponents struct {
							Text      string `xml:",chardata"`
							Innertube []struct {
								Text        string `xml:",chardata"`
								Name        string `xml:"name"`
								ID          string `xml:"id"`
								Axialoffset struct {
									Text   string `xml:",chardata"`
									Method string `xml:"method,attr"`
								} `xml:"axialoffset"`
								Position struct {
									Text string `xml:",chardata"`
									Type string `xml:"type,attr"`
								} `xml:"position"`
								Overridemass              string `xml:"overridemass"`
								Overridesubcomponentsmass string `xml:"overridesubcomponentsmass"`
								Material                  struct {
									Text    string `xml:",chardata"`
									Type    string `xml:"type,attr"`
									Density string `xml:"density,attr"`
								} `xml:"material"`
								Length               string `xml:"length"`
								Radialposition       string `xml:"radialposition"`
								Radialdirection      string `xml:"radialdirection"`
								Outerradius          string `xml:"outerradius"`
								Thickness            string `xml:"thickness"`
								Clusterconfiguration string `xml:"clusterconfiguration"`
								Clusterscale         string `xml:"clusterscale"`
								Clusterrotation      string `xml:"clusterrotation"`
								Subcomponents        struct {
									Text          string `xml:",chardata"`
									Masscomponent struct {
										Text        string `xml:",chardata"`
										Name        string `xml:"name"`
										ID          string `xml:"id"`
										Axialoffset struct {
											Text   string `xml:",chardata"`
											Method string `xml:"method,attr"`
										} `xml:"axialoffset"`
										Position struct {
											Text string `xml:",chardata"`
											Type string `xml:"type,attr"`
										} `xml:"position"`
										Packedlength      string `xml:"packedlength"`
										Packedradius      string `xml:"packedradius"`
										Radialposition    string `xml:"radialposition"`
										Radialdirection   string `xml:"radialdirection"`
										Mass              string `xml:"mass"`
										Masscomponenttype string `xml:"masscomponenttype"`
									} `xml:"masscomponent"`
									Centeringring []struct {
										Text               string `xml:",chardata"`
										Name               string `xml:"name"`
										ID                 string `xml:"id"`
										Instancecount      string `xml:"instancecount"`
										Instanceseparation string `xml:"instanceseparation"`
										Axialoffset        struct {
											Text   string `xml:",chardata"`
											Method string `xml:"method,attr"`
										} `xml:"axialoffset"`
										Position struct {
											Text string `xml:",chardata"`
											Type string `xml:"type,attr"`
										} `xml:"position"`
										Material struct {
											Text    string `xml:",chardata"`
											Type    string `xml:"type,attr"`
											Density string `xml:"density,attr"`
										} `xml:"material"`
										Length          string `xml:"length"`
										Radialposition  string `xml:"radialposition"`
										Radialdirection string `xml:"radialdirection"`
										Outerradius     string `xml:"outerradius"`
										Innerradius     string `xml:"innerradius"`
									} `xml:"centeringring"`
									Bulkhead struct {
										Text               string `xml:",chardata"`
										Name               string `xml:"name"`
										ID                 string `xml:"id"`
										Instancecount      string `xml:"instancecount"`
										Instanceseparation string `xml:"instanceseparation"`
										Axialoffset        struct {
											Text   string `xml:",chardata"`
											Method string `xml:"method,attr"`
										} `xml:"axialoffset"`
										Position struct {
											Text string `xml:",chardata"`
											Type string `xml:"type,attr"`
										} `xml:"position"`
										Material struct {
											Text    string `xml:",chardata"`
											Type    string `xml:"type,attr"`
											Density string `xml:"density,attr"`
										} `xml:"material"`
										Length          string `xml:"length"`
										Radialposition  string `xml:"radialposition"`
										Radialdirection string `xml:"radialdirection"`
										Outerradius     string `xml:"outerradius"`
									} `xml:"bulkhead"`
								} `xml:"subcomponents"`
								Motormount struct {
									Text          string `xml:",chardata"`
									Ignitionevent string `xml:"ignitionevent"`
									Ignitiondelay string `xml:"ignitiondelay"`
									Overhang      string `xml:"overhang"`
									Motor         []struct {
										Text         string `xml:",chardata"`
										Configid     string `xml:"configid,attr"`
										Type         string `xml:"type"`
										Manufacturer string `xml:"manufacturer"`
										Digest       string `xml:"digest"`
										Designation  string `xml:"designation"`
										Diameter     string `xml:"diameter"`
										Length       string `xml:"length"`
										Delay        string `xml:"delay"`
									} `xml:"motor"`
									Ignitionconfiguration []struct {
										Text          string `xml:",chardata"`
										Configid      string `xml:"configid,attr"`
										Ignitionevent string `xml:"ignitionevent"`
										Ignitiondelay string `xml:"ignitiondelay"`
									} `xml:"ignitionconfiguration"`
								} `xml:"motormount"`
							} `xml:"innertube"`
							Trapezoidfinset struct {
								Text       string `xml:",chardata"`
								Name       string `xml:"name"`
								ID         string `xml:"id"`
								Appearance struct {
									Text  string `xml:",chardata"`
									Paint struct {
										Text  string `xml:",chardata"`
										Red   string `xml:"red,attr"`
										Green string `xml:"green,attr"`
										Blue  string `xml:"blue,attr"`
										Alpha string `xml:"alpha,attr"`
									} `xml:"paint"`
									Shine string `xml:"shine"`
								} `xml:"appearance"`
								Instancecount string `xml:"instancecount"`
								Fincount      string `xml:"fincount"`
								Radiusoffset  struct {
									Text   string `xml:",chardata"`
									Method string `xml:"method,attr"`
								} `xml:"radiusoffset"`
								Angleoffset struct {
									Text   string `xml:",chardata"`
									Method string `xml:"method,attr"`
								} `xml:"angleoffset"`
								Rotation    string `xml:"rotation"`
								Axialoffset struct {
									Text   string `xml:",chardata"`
									Method string `xml:"method,attr"`
								} `xml:"axialoffset"`
								Position struct {
									Text string `xml:",chardata"`
									Type string `xml:"type,attr"`
								} `xml:"position"`
								Finish   string `xml:"finish"`
								Material struct {
									Text    string `xml:",chardata"`
									Type    string `xml:"type,attr"`
									Density string `xml:"density,attr"`
								} `xml:"material"`
								Thickness    string `xml:"thickness"`
								Crosssection string `xml:"crosssection"`
								Cant         string `xml:"cant"`
								Tabheight    string `xml:"tabheight"`
								Tablength    string `xml:"tablength"`
								Tabposition  []struct {
									Text       string `xml:",chardata"`
									Relativeto string `xml:"relativeto,attr"`
								} `xml:"tabposition"`
								Filletradius   string `xml:"filletradius"`
								Filletmaterial struct {
									Text    string `xml:",chardata"`
									Type    string `xml:"type,attr"`
									Density string `xml:"density,attr"`
								} `xml:"filletmaterial"`
								Rootchord   string `xml:"rootchord"`
								Tipchord    string `xml:"tipchord"`
								Sweeplength string `xml:"sweeplength"`
								Height      string `xml:"height"`
							} `xml:"trapezoidfinset"`
						} `xml:"subcomponents"`
					} `xml:"bodytube"`
				} `xml:"subcomponents"`
			} `xml:"stage"`
		} `xml:"subcomponents"`
	} `xml:"rocket"`
	Simulations struct {
		Text       string `xml:",chardata"`
		Simulation []struct {
			Text       string `xml:",chardata"`
			Status     string `xml:"status,attr"`
			Name       string `xml:"name"`
			Simulator  string `xml:"simulator"`
			Calculator string `xml:"calculator"`
			Conditions struct {
				Text               string `xml:",chardata"`
				Configid           string `xml:"configid"`
				Launchrodlength    string `xml:"launchrodlength"`
				Launchrodangle     string `xml:"launchrodangle"`
				Launchroddirection string `xml:"launchroddirection"`
				Windaverage        string `xml:"windaverage"`
				Windturbulence     string `xml:"windturbulence"`
				Launchaltitude     string `xml:"launchaltitude"`
				Launchlatitude     string `xml:"launchlatitude"`
				Launchlongitude    string `xml:"launchlongitude"`
				Geodeticmethod     string `xml:"geodeticmethod"`
				Atmosphere         struct {
					Text  string `xml:",chardata"`
					Model string `xml:"model,attr"`
				} `xml:"atmosphere"`
				Timestep string `xml:"timestep"`
			} `xml:"conditions"`
			Flightdata struct {
				Text               string   `xml:",chardata"`
				Maxaltitude        string   `xml:"maxaltitude,attr"`
				Maxvelocity        string   `xml:"maxvelocity,attr"`
				Maxacceleration    string   `xml:"maxacceleration,attr"`
				Maxmach            string   `xml:"maxmach,attr"`
				Timetoapogee       string   `xml:"timetoapogee,attr"`
				Flighttime         string   `xml:"flighttime,attr"`
				Groundhitvelocity  string   `xml:"groundhitvelocity,attr"`
				Launchrodvelocity  string   `xml:"launchrodvelocity,attr"`
				Optimumdelay       string   `xml:"optimumdelay,attr"`
				Deploymentvelocity string   `xml:"deploymentvelocity,attr"`
				Warning            []string `xml:"warning"`
			} `xml:"flightdata"`
		} `xml:"simulation"`
	} `xml:"simulations"`
	Photostudio string `xml:"photostudio"`
}
