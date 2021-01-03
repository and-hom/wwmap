export interface Region {
    id: bigint,
    title: string,
}

export interface RegionWithRivers extends Region {
    rivers: River[],
}


export interface River {
    id: bigint,
    title: string,
}

export interface RiverFull extends River {
    description: string,
    reports: ReportGroup[],
}

export interface ReportGroup {
    source: string,
    reports: Report[],
}

export interface Report {
    source: string,
    id: string,
    title: string,
    url: string,
    source_logo_url: string,
}