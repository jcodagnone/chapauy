import { OffensesResponse, Facet, SidebarMode, Dimension } from '../../lib/types';

export const mockFacets: Facet[] = [
    {
        dimension: Dimension.Database,
        total_values: 2,
        values: [
            { value: 'CGM', label: 'CGM', count: 100, selected: false },
            { value: 'SUCIVE', label: 'SUCIVE', count: 50, selected: false },
        ],
    },
    {
        dimension: Dimension.Year,
        total_values: 5,
        values: [
            { value: '2024', label: '2024', count: 80, selected: false },
            { value: '2023', label: '2023', count: 70, selected: false },
        ],
    },
];

export const mockOffenses: OffensesResponse = {
    offenses: [
        {
            id: '1',
            doc_id: 'doc1',
            doc_date: '2024-01-01',
            doc_source: 'https://example.com/doc1',
            country: 'UY',
            adm_division: 'Montevideo',
            vehicle_type: 'Auto',
            vehicle: 'SGB1234',
            description: 'Speeding',
            location: 'Av Italia',
            time: '12:00',
            mercosur_format: true,
            repo_id: 1,
            record_id: 1,
            ur: 1,
            point: { lat: -34.9, lng: -56.1 },
        },
        {
            id: '2',
            doc_id: 'doc2',
            doc_date: '2024-01-02',
            doc_source: 'https://example.com/doc2',
            country: 'UY',
            adm_division: 'Montevideo',
            vehicle_type: 'Moto',
            vehicle: 'SGB5678',
            description: 'Parking',
            location: '18 de Julio',
            time: '14:00',
            mercosur_format: true,
            repo_id: 1,
            record_id: 2,
            ur: 1,
            point: { lat: -34.9, lng: -56.1 },
        }
    ],
    pagination: {
        current_page: 1,
        total_pages: 5,
    },
    repos: {
        '1': { name: 'Montevideo' }
    },
    summary: {
        avg_ur: 1.5,
        facets: mockFacets,
        record_count: 150,
        total_ur: 200,
    },
    chartData: {
        dayOfWeek: { 'Monday': { 'count': 10 } },
        dayOfYear: { '2024-01-01': { 'count': 5 } },
        timeOfDay: { '12': { 'count': 8 } }
    }
};

export const mockSidebarResponse = mockFacets;
