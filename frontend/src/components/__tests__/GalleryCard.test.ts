import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import GalleryCard from '../GalleryCard.vue'

describe('GalleryCard', () => {
  const mockImage = {
    id: 1,
    title: 'Test Image',
    lsky_url: 'https://example.com/test.png',
    tags: ['nature', 'travel'],
    uploaded_by: 42,
    created_at: '2025-01-15T10:00:00Z',
  }

  it('renders image title', () => {
    const wrapper = mount(GalleryCard, {
      props: { image: mockImage },
    })
    expect(wrapper.text()).toContain('Test Image')
  })

  it('renders image tags', () => {
    const wrapper = mount(GalleryCard, {
      props: { image: mockImage },
    })
    expect(wrapper.text()).toContain('nature')
    expect(wrapper.text()).toContain('travel')
  })

  it('renders the image with correct src', () => {
    const wrapper = mount(GalleryCard, {
      props: { image: mockImage },
    })
    const img = wrapper.find('img')
    expect(img.attributes('src')).toBe('https://example.com/test.png')
    expect(img.attributes('alt')).toBe('Test Image')
  })

  it('emits click event with image id when clicked', () => {
    const wrapper = mount(GalleryCard, {
      props: { image: mockImage },
    })
    wrapper.trigger('click')
    expect(wrapper.emitted('click')).toBeTruthy()
    expect(wrapper.emitted('click')![0]).toEqual([1])
  })

  it('renders formatted date', () => {
    const wrapper = mount(GalleryCard, {
      props: { image: mockImage },
    })
    // The date should be rendered somewhere in the card
    expect(wrapper.text()).toContain('2025')
  })

  it('shows fallback text on image error', async () => {
    const wrapper = mount(GalleryCard, {
      props: { image: mockImage },
    })
    const img = wrapper.find('img')
    await img.trigger('error')
    expect(wrapper.text()).toContain('加载失败')
  })
})
